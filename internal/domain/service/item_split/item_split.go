package item_split

import log "github.com/sirupsen/logrus"

type FileSplitService struct {
	l *log.Entry
}

func NewFileSplitService(l *log.Logger) *FileSplitService {
	return &FileSplitService{
		l: l.WithField("component", "FileSplitService"),
	}
}

func (s *FileSplitService) SplitFileBySize(size int64, maxParts int) []int64 {
	if size/int64(maxParts) == 0 {
		return []int64{0}
	}

	chunkSize := size / int64(maxParts)

	res := make([]int64, maxParts, maxParts)

	var i int64
	for i = 0; i < int64(maxParts); i++ {
		res[i] = chunkSize * i
	}

	return res
}
