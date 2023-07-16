package item_split_service

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type FileSplitService struct {
	l *log.Entry
}

func NewFileSplitService(l *log.Logger) *FileSplitService {
	return &FileSplitService{
		l: l.WithField("component", "FileSplitService"),
	}
}

func (s *FileSplitService) SplitFileBySize(size int64, maxParts int) ([]int64, error) {
	if maxParts < 1 {
		return nil, errors.Errorf("incorrect parts number: %d", maxParts)
	}

	if size/int64(maxParts) == 0 {
		return []int64{0}, nil
	}

	chunkSize := size / int64(maxParts)

	res := make([]int64, maxParts, maxParts)

	var i int64
	for i = 0; i < int64(maxParts); i++ {
		res[i] = chunkSize * i
	}

	return res, nil
}
