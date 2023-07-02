package file_server

import "io"

type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}
