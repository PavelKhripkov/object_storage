package item_usecase

import (
	"context"
	file_server2 "github.com/PavelKhripkov/object_storage/internal/domain/model/file_server"
	item2 "github.com/PavelKhripkov/object_storage/internal/domain/model/item"
	chunk2 "github.com/PavelKhripkov/object_storage/internal/domain/service/chunk"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/file_server"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/item"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/item_split"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
)

const defaultPartsCount = 6 // TODO better get from config

type Usecase struct {
	itemService       *item.Service
	chunkService      *chunk2.Service
	fileServerService *file_server.Service
	fileSplitService  *item_split.FileSplitService

	l *log.Entry
}

func NewItemUsecase(
	itemService *item.Service,
	chunkService *chunk2.Service,
	fileService *file_server.Service,
	fileSplitService *item_split.FileSplitService,
	l *log.Logger) *Usecase {
	return &Usecase{
		itemService:       itemService,
		chunkService:      chunkService,
		fileServerService: fileService,
		fileSplitService:  fileSplitService,
		l:                 l.WithField("component", "ItemUsecase"),
	}
}

func (s *Usecase) Get(ctx context.Context, id string) (item2.Item, error) {
	res, err := s.itemService.Get(ctx, id)
	if err != nil {
		return item2.Item{}, err
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

func (s *Usecase) Store(ctx context.Context, dto StoreItemDTO) (item2.Item, error) {
	params := item.CreateItemDTO{
		Name:     dto.Name,
		Path:     dto.Path,
		Size:     dto.Size,
		BucketId: dto.BucketID,
	}

	newItem, err := s.itemService.Create(ctx, params)
	if err != nil {
		return item2.Item{}, err
	}

	go s.store(context.TODO(), newItem, dto)

	return newItem, nil
}

func (s *Usecase) store(ctx context.Context, itm item2.Item, dto StoreItemDTO) {
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
	usedServices := make(map[string]file_server2.FileServer)

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
				createItemParams := chunk2.CreateChunkDTO{
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

	changeItemParams := item.ChangeItemDTO{
		Status:     item2.ItemStatusOK.Pointer(),
		ChunkCount: &chunkPosCount,
	}

	_, err = s.itemService.Update(ctx, itm, changeItemParams)
	if err != nil {
		// TODO handle error
	}
}

func (s *Usecase) storeWorker(ctx context.Context, f file_server.Opener, c chunk, fileService file_server2.FileServer, queue chan<- chunk) {
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

type part struct {
	start, end int64
	file       io.ReadSeeker
}

type ContentMapper struct {
	io.ReadSeeker
	parts []*part
	size  int64

	mu     sync.Mutex
	offset int64
}

func (s *ContentMapper) Seek(offset int64, whence int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset += s.offset
	case io.SeekEnd:
		offset += s.size
	default:
		return s.offset, errors.Errorf("unknown whence %d", whence)
	}

	if offset < 0 {
		return s.offset, os.ErrInvalid
	}

	s.offset = offset
	return s.offset, nil
}

func (s *ContentMapper) Read(b []byte) (int, error) {
	var read int
	curr := s.findPart()

	if curr == nil {
		return 0, io.EOF
	}

	for read < len(b) {
		rb := b[read:]
		n, err := curr.Read(rb)
		read += n
		s.mu.Lock()
		s.offset += int64(n)
		s.mu.Unlock()
		if err == io.EOF {
			curr = s.findPart()
			if curr == nil {
				return read, io.EOF
			}
		} else if err != nil {
			return read, err
		}
	}

	return read, nil
}

func (s *ContentMapper) findPart() io.ReadSeeker {
	for _, part := range s.parts {
		if s.offset >= part.start && s.offset <= part.end {
			_, err := part.file.Seek(s.offset-part.start, io.SeekStart)
			if err != nil {
				return nil
			}
			return part.file
		}
	}
	return nil
}

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

	parts := make([]*part, len(chunks))
	var nextStart int64

	for i, chnk := range chunks {
		chunkFile, err := s.fileServerService.OpenChunkFile(ctx, chnk)
		if err != nil {
			return nil, "", err
		}
		newPart := part{
			start: nextStart,
			end:   nextStart + chnk.Size - 1,
			file:  chunkFile,
		}
		nextStart += chnk.Size
		parts[i] = &newPart
	}

	contentMapper := &ContentMapper{
		parts: parts,
		size:  itm.Size,
	}

	return contentMapper, itm.Name, nil
}
