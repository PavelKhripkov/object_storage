package item_usecase

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/file_server"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/item"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/policy"
	"github.com/bufbuild/buf/private/pkg/uuidutil"
	"sync"
)

type Usecase struct {
	itemService      item.Service
	fileService      fileService
	fileSplitService policy.FileSplitService
}

func NewItemUsecase(s item.Service) *Usecase {
	return &Usecase{itemService: s}
}

func (s *Usecase) UploadItem(ctx context.Context, f file_server.File, name string, size int64) (string, error) {
	chunkPositions := s.fileSplitService.SplitFileBySize(size)

	chunks := make([]chunk, len(chunkPositions), len(chunkPositions))

	for i, c := range chunkPositions {
		id, err := uuidutil.New()
		if err != nil {
			return "", err
		}
		newChunk := chunk{ID: id, Start: c}
		if i != len(chunks)-1 {
			newChunk.End = chunkPositions[i+1] - 1
		} else {
			newChunk.End = size - 1
		}
	}

	go s.store(context.TODO(), f, size, chunks)

	return "", nil
}

type chunk struct {
	ID            string
	Start, End    int64
	Stored        bool
	FileServiceID string
}

func (s *Usecase) store(ctx context.Context, f file_server.File, size int64, chunks []chunk) {
	chunkChannel := make(chan chunk, len(chunks))
	defer close(chunkChannel)

	for _, c := range chunks {
		chunkChannel <- c
	}

	success := 0
	usedServices := make(map[string]fileService)

	for success < len(chunks) {
		select {
		case c := <-chunkChannel:
			if !c.Stored {
				delete(usedServices, c.FileServiceID)

				fileService, err := s.fileService.ChooseOne(ctx, usedServices)
				if err != nil {
					// TODO handle error
				}

				usedServices[fileService.ID()] = fileService
				go s.worker(ctx, f, c, fileService, chunkChannel)

			} else {
				// save info to db
			}
		case <-ctx.Done():
		}

	}

	wg := sync.WaitGroup{}
	wg.Add(len(chunks))

}

func (s *Usecase) worker(ctx context.Context, f file_server.File, c chunk, fileService fileService, queue chan<- chunk) {
	if err := fileService.Store(ctx, f, c.Start, c.End-c.Start+1); err == nil {
		c.Stored = true
	} else {
		// TODO log error
	}

	c.FileServiceID = fileService.ID()
	queue <- c
}
