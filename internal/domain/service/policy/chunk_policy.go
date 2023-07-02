package policy

type FileSplitService struct {
}

func NewFileSplitService() *FileSplitService {
	return &FileSplitService{}
}

func (s *FileSplitService) SplitFileBySize(size int64) []int64 {
	if size/6 == 0 {
		return []int64{0}
	}

	chunkSize := size / 6

	res := make([]int64, 6, 6)

	var i int64
	for i = 0; i < 6; i++ {
		res[i] = chunkSize * i
	}

	return res
}
