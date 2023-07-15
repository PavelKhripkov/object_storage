package item_usecase

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/file_server_model"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/item_model"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/chunk_service"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/file_server_service"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/item_service"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/item_split_service"
	"github.com/PavelKhripkov/object_storage/pkg/content_mapper"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
)

const defaultPartsCount = 6 // TODO better get from config

type Usecase struct {
	itemService       *item_service.Service
	chunkService      *chunk_service.Service
	fileServerService *file_server_service.Service
	fileSplitService  *item_split_service.FileSplitService

	l *log.Entry
}

func NewItemUsecase(
	itemService *item_service.Service,
	chunkService *chunk_service.Service,
	fileService *file_server_service.Service,
	fileSplitService *item_split_service.FileSplitService,
	l *log.Logger) *Usecase {
	return &Usecase{
		itemService:       itemService,
		chunkService:      chunkService,
		fileServerService: fileService,
		fileSplitService:  fileSplitService,
		l:                 l.WithField("component", "itemUsecase"),
	}
}

func (s *Usecase) Get(ctx context.Context, id string) (item_model.Item, error) {
	res, err := s.itemService.Get(ctx, id)
	if err != nil {
		return item_model.Item{}, err
	}

	return res, nil
}

type chunk struct {
	Position      uint8
	Start, End    int64
	Processed     bool
	FileServiceID string
	FilePath      string
}

func (s *Usecase) Store(ctx context.Context, dto StoreItemDTO) (item_model.Item, error) {
	params := item_service.CreateItemDTO{
		Name:        dto.Name,
		ContainerID: dto.ContainerID,
		Size:        dto.Size,
	}

	newItem, err := s.itemService.Create(ctx, params)
	if err != nil {
		return item_model.Item{}, err
	}

	go s.store(context.TODO(), newItem, dto)

	return newItem, nil
}

func (s *Usecase) store(ctx context.Context, itm item_model.Item, dto StoreItemDTO) {
	partsCount, err := s.fileServerService.Count(ctx)
	if err != nil {
		// TODO handle error
		return
	}

	if partsCount > defaultPartsCount {
		partsCount = defaultPartsCount
	}

	chunkPositions := s.fileSplitService.SplitFileBySize(dto.Size, partsCount)
	chunkPosCount := uint8(len(chunkPositions))
	chunks := make([]chunk, chunkPosCount)

	for i, c := range chunkPositions {
		newChunk := chunk{
			Position: uint8(i),
			Start:    c,
		}

		if i != len(chunks)-1 {
			newChunk.End = chunkPositions[i+1] - 1
		} else {
			newChunk.End = dto.Size - 1
		}

		chunks[i] = newChunk
	}

	chunkChannel := make(chan chunk, len(chunks))
	defer close(chunkChannel)

	for _, c := range chunks {
		chunkChannel <- c
	}

	success := 0
	usedServices := make(map[string]file_server_model.FileServer)

	for success < len(chunks) {
		select {
		case c := <-chunkChannel:
			if !c.Processed {
				delete(usedServices, c.FileServiceID)

				fileServer, err := s.fileServerService.ChooseOneExcluding(ctx, usedServices)
				if err != nil {
					s.l.WithError(err).Error("File server selection error.")
				}

				usedServices[fileServer.GetID()] = fileServer
				go s.storeWorker(ctx, dto.F, c, fileServer, chunkChannel)

			} else {
				createItemParams := chunk_service.CreateChunkDTO{
					ItemID:       itm.ID,
					FileServerID: c.FileServiceID,
					FilePath:     c.FilePath,
					Position:     c.Position,
					Size:         c.End - c.Start + 1,
				}

				_, err = s.chunkService.Create(ctx, createItemParams)
				if err != nil {
					// TODO handle error
				}
				success++
			}
		case <-ctx.Done():
			// TODO clean up created chunks
		}

	}

	changeItemParams := item_service.ChangeItemDTO{
		Status:     item_model.ItemStatusOK.Pointer(),
		ChunkCount: &chunkPosCount,
	}

	_, err = s.itemService.Update(ctx, itm, changeItemParams)
	if err != nil {
		// TODO handle error
	}
}

func (s *Usecase) storeWorker(ctx context.Context, f file_server_service.Opener, c chunk, fileService file_server_model.FileServer, queue chan<- chunk) {
	if FilePath, err := s.fileServerService.StoreChunk(ctx, fileService, f, c.Start, c.End-c.Start+1); err != nil {
		s.l.Error(err)
		return
	} else {

		c.FilePath = FilePath
		c.Processed = true
	}

	c.FileServiceID = fileService.GetID()
	queue <- c
}

//func (s *Usecase) Download(ctx context.Context, id string) (string, string, error) {
//	itm, err := s.itemService.Get(ctx, id)
//	if err != nil {
//		return "", "", err
//	}
//
//	chunks, err := s.chunkService.GetItemChunks(ctx, id)
//	if err != nil {
//		return "", "", err
//	}
//
//	if len(chunks) != int(itm.ChunkCount) {
//		return "", "", errors.New("wrong chunk amount")
//	}
//
//	localFile, err := os.CreateTemp("", "object_storage_tmp_file")
//	if err != nil {
//		return "", "", err
//	}
//	localFile.Close()
//
//	tempFileName := localFile.Name()
//
//	var nextPos int64
//
//	wg := sync.WaitGroup{}
//	wg.Add(int(itm.ChunkCount))
//
//	for _, chnk := range chunks {
//		pos := nextPos
//		currChnk := chnk
//		helper := func() {
//			defer wg.Done()
//			s.fileServerService.CopyChunkToLocal(ctx, currChnk, pos, tempFileName)
//		}
//		go helper()
//		nextPos += chnk.Size
//	}
//
//	wg.Wait()
//
//	return tempFileName, itm.Name, nil
//
//}

func (s *Usecase) Download(ctx context.Context, id string) (io.ReadSeeker, string, error) {
	itm, err := s.itemService.Get(ctx, id)
	if err != nil {
		return nil, "", err
	}

	chunks, err := s.chunkService.GetItemChunks(ctx, id)
	if err != nil {
		return nil, "", err
	}

	if len(chunks) != int(itm.ChunkCount) {
		return nil, "", errors.New("wrong chunk amount")
	}

	parts := make([]*content_mapper.Part, len(chunks))
	var nextStart int64

	for i, chnk := range chunks {
		chunkFile, err := s.fileServerService.OpenChunkFile(ctx, chnk)
		if err != nil {
			return nil, "", err
		}
		newPart := content_mapper.Part{
			Start: nextStart,
			End:   nextStart + chnk.Size - 1,
			Open:  chunkFile,
		}
		nextStart += chnk.Size
		parts[i] = &newPart
	}

	contentMapper, err := content_mapper.NewContentMapper(s.l.Logger, parts, itm.Size)
	if err != nil {
		return nil, "", err
	}

	return contentMapper, itm.Name, nil
}
