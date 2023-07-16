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
	"mime/multipart"
)

const defaultPartsCount = 6 // TODO better get from config

// Usecase represents item use cases.
type Usecase struct {
	itemService       *item_service.Service
	chunkService      *chunk_service.Service
	fileServerService *file_server_service.Service
	fileSplitService  *item_split_service.FileSplitService

	l *log.Entry
}

// NewItemUsecase creates new item use cases service.
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

// Get returns item model.
func (s *Usecase) Get(ctx context.Context, id string) (item_model.Item, error) {
	res, err := s.itemService.Get(ctx, id)
	if err != nil {
		return item_model.Item{}, err
	}

	return res, nil
}

type chunkJob struct {
	Position      uint8
	Start, End    int64
	Processed     bool
	FileServiceID string
	FilePath      string
}

// Store creates item model and starts storing item chunks on file servers.
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

// store runs in background and performs:
// 1. item splitting into chunks;
// 2. getting available file servers;
// 3. storing item chunks on file servers.
func (s *Usecase) store(ctx context.Context, itm item_model.Item, dto StoreItemDTO) {
	s.l.Infof("Storing file %s, of size %d bytes.", dto.Name, dto.Size)

	if dto.Close != nil {
		defer dto.Close()
	}

	partsCount := defaultPartsCount

	fileServerCount, err := s.fileServerService.Count(ctx)
	if err != nil {
		// TODO handle error
		return
	}

	if fileServerCount < 1 {
		s.l.Error("No available file servers found.")
		_, err := s.itemService.Update(ctx, itm, item_service.UpdateItemDTO{Status: item_model.ItemStatusFail.Pointer()})
		if err != nil {
			s.l.Error(err)
		}
		return
	}

	// If there are too few available file servers, we're reducing target chunk amount.
	if partsCount > fileServerCount {
		partsCount = fileServerCount
	}

	// Split item into chunks.
	chunkPositions, err := s.fileSplitService.SplitFileBySize(dto.Size, partsCount)
	if err != nil {
		s.l.Error(err)
		_, err = s.itemService.Update(ctx, itm, item_service.UpdateItemDTO{Status: item_model.ItemStatusFail.Pointer()})
		if err != nil {
			s.l.Error(err)
		}
		return
	}

	chunkJobs := make([]chunkJob, partsCount)

	// Create chunk jobs.
	for i, c := range chunkPositions {
		newChunkJob := chunkJob{
			Position: uint8(i),
			Start:    c,
		}

		if i != len(chunkJobs)-1 {
			newChunkJob.End = chunkPositions[i+1] - 1
		} else {
			newChunkJob.End = dto.Size - 1
		}

		chunkJobs[i] = newChunkJob
	}

	jobChannel := make(chan chunkJob, len(chunkJobs))
	defer close(jobChannel)

	// Since jobs are small, we can store all of them in a buffered channel.
	for _, c := range chunkJobs {
		jobChannel <- c
	}

	success := 0
	usedServices := make(map[string]bool)

	// Reading from job channel until either all chunks are stored successfully or unrecoverable error encountered.
	for success < len(chunkJobs) {
		select {
		case c := <-jobChannel:
			// New job or the one that couldn't be stored on a file server.
			if !c.Processed {
				delete(usedServices, c.FileServiceID)

				fileServer, err := s.fileServerService.ChooseOneExcluding(ctx, usedServices)
				if err != nil {
					s.l.WithError(err).Error("File server selection error.")
				}

				if fileServer.GetFreeSpace() < c.End-c.Start+1 {
					s.l.Error("no free space on file servers")

					_, err = s.itemService.Update(ctx, itm, item_service.UpdateItemDTO{Status: item_model.ItemStatusFail.Pointer()})
					if err != nil {
						s.l.Error(err)
					}
					return

					// TODO clean up created chunks
				}

				usedServices[fileServer.GetID()] = true
				go s.storeWorker(ctx, dto.F, c, fileServer, jobChannel)

				// The one successfully stored.
			} else {
				createParams := chunk_service.CreateChunkDTO{
					ItemID:       itm.ID,
					FileServerID: c.FileServiceID,
					FilePath:     c.FilePath,
					Position:     c.Position,
					Size:         c.End - c.Start + 1,
				}

				_, err = s.chunkService.Create(ctx, createParams)
				if err != nil {
					// TODO handle error
				}

				err = s.fileServerService.UpdateUsedSpace(ctx, c.FileServiceID, createParams.Size)
				if err != nil {
					return
				}

				success++
			}
		case <-ctx.Done():
			s.l.Warn(ctx.Err())
			// TODO clean up created chunkJobs
		}

	}
	chunkPosCount := uint8(len(chunkPositions))

	// Now can store chunk info into item model.
	changeItemParams := item_service.UpdateItemDTO{
		Status:     item_model.ItemStatusOK.Pointer(),
		ChunkCount: &chunkPosCount,
	}

	_, err = s.itemService.Update(ctx, itm, changeItemParams)
	if err != nil {
		s.l.Error(err)
	}
}

// storeWorker stores chunk on file server and replies into job queue with results.
func (s *Usecase) storeWorker(ctx context.Context, f *multipart.FileHeader, c chunkJob, fileService file_server_model.FileServer, queue chan<- chunkJob) {
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

// Download prepares chunks, opens streams and returns io.ReadSeeker that can be used to read item seamlessly.
func (s *Usecase) Download(ctx context.Context, id string) (io.ReadSeeker, string, error) {
	itm, err := s.itemService.Get(ctx, id)
	if err != nil {
		return nil, "", err
	}

	// TODO item can be taken from local server until it's not transferred to remote ones.
	if itm.Status != item_model.ItemStatusOK {
		return nil, "", errors.Errorf("item can't be downloaded, current status: %s", itm.Status)
	}

	chunks, err := s.chunkService.GetItemChunks(ctx, id)
	if err != nil {
		return nil, "", err
	}

	if len(chunks) != int(itm.ChunkCount) {
		return nil, "", errors.New("wrong chunkJob amount")
	}

	parts := make([]*content_mapper.Part, len(chunks))
	var nextStart int64

	// Preparing chunk files for content mapper.
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

	contentMapper, err := content_mapper.NewContentMapper(s.l.Logger.WithField("component", "ContentMapper"), parts, itm.Size)
	if err != nil {
		return nil, "", err
	}

	return contentMapper, itm.Name, nil
}
