package content_mapper

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
)

// Part represents one part of a stream to be read.
// TODO need constructor?
type Part struct {
	Start, End int64
	Open       func() (io.ReadSeekCloser, error)
	close      func() error
	file       io.ReadSeeker
}

// ContentMapper maps reading of a single source into multiple parts.
// Handles Open, Close, Seek, Read methods of Parts.
type ContentMapper struct {
	io.ReadSeeker
	parts      []*Part
	size       int64
	currentPos int
	l          *logrus.Entry

	mu     sync.Mutex
	offset int64
}

// NewContentMapper create new content mapper with provided parts.
func NewContentMapper(l *logrus.Entry, parts []*Part, size int64) (*ContentMapper, error) {
	res := &ContentMapper{
		l:     l,
		parts: parts,
		size:  size,
	}

	if err := res.validate(); err != nil {
		return nil, err
	}

	return res, nil
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
	s.mu.Lock()
	defer s.mu.Unlock()

	var read int
	var err error
	curr := s.parts[s.currentPos].file

	if curr == nil {
		curr, err = s.open()
		if err != nil {
			return 0, err
		}
	}

	for read < len(b) {
		rb := b[read:]
		n, err := curr.Read(rb)
		read += n
		s.offset += int64(n)

		if err == io.EOF {
			curr, err = s.next()
			if err != nil {
				return read, err
			}
		} else if err != nil {
			return read, err
		}
	}

	return read, nil
}

func (s *ContentMapper) open() (io.ReadSeeker, error) {
	for i, part := range s.parts {
		if s.offset >= part.Start && s.offset <= part.End {
			file, err := part.Open()
			if err != nil {
				return nil, err
			}

			part.close = func() error {
				return file.Close()
			}

			_, err = file.Seek(s.offset-part.Start, io.SeekStart)
			if err != nil {
				return nil, err
			}

			s.currentPos = i
			s.parts[s.currentPos].file = file

			return file, nil
		}
	}
	return nil, io.ErrUnexpectedEOF
}

func (s *ContentMapper) next() (io.ReadSeeker, error) {
	if s.parts[s.currentPos].close != nil {
		s.parts[s.currentPos].file = nil
		err := s.parts[s.currentPos].close()
		if err != nil {
			return nil, err
		}
	}

	if s.currentPos < len(s.parts)-1 {
		s.currentPos++

		file, err := s.parts[s.currentPos].Open()
		if err != nil {
			return nil, err
		}

		s.parts[s.currentPos].file = file

		s.parts[s.currentPos].close = func() error {
			return file.Close()
		}

		return file, nil
	}

	return nil, io.EOF
}

func (s *ContentMapper) validate() error {
	// TODO validate
	return nil
}
