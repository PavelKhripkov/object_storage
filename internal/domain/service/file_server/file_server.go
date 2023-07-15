package file_server

import (
	"context"
	"encoding/json"
	"github.com/PavelKhripkov/object_storage/internal/adapter/db/sqlite"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/chunk"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/file_server"
	"github.com/PavelKhripkov/object_storage/pkg/client/ssh"
	"github.com/gofrs/uuid/v5"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"strconv"
	"time"
)

type Service struct {
	storage fileServerStorage
	l       *log.Entry
}

func NewFileServerService(fileServerStorage fileServerStorage, l *log.Logger) *Service {
	return &Service{
		storage: fileServerStorage,
		l:       l.WithField("component", "FileServerService"),
	}
}

func (s Service) ChooseOneExcluding(ctx context.Context, exclude map[string]file_server.FileServer) (file_server.FileServer, error) {
	i := 0
	excludeList := make([]string, len(exclude))
	for k := range exclude {
		excludeList[i] = k
		i++
	}

	common, err := s.storage.ChooseOneExcluding(ctx, excludeList)
	if err != nil {
		return nil, err
	}

	res, err := s.fromCommonDTO(common)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s Service) Ping(fs file_server.FileServer) error {
	switch server := fs.(type) {
	case *file_server.SSHFileServer:
		_, closeFunc, err := ssh.NewClient(server.Address, server.Port, server.User, server.Key)
		if err != nil {
			return err
		}
		if err := closeFunc(); err != nil {
			s.l.Error(err)
		}
	case *file_server.APIFileServer:
		return nil

	}

	return nil
}

func (s Service) Add(ctx context.Context, dto AddFileServerDTO) (file_server.FileServer, error) {
	commonServer := sqlite.CommonFileServerDTO{
		Name: dto.GetName(),
		Type: dto.GetType(),
	}

	params, err := dto.MarshalParams()
	if err != nil {
		return nil, err
	}

	commonServer.Params = params

	newID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	commonServer.ID = newID.String()

	now := time.Now()

	commonServer.Created = now
	commonServer.Modified = now

	if err := s.storage.Add(ctx, commonServer); err != nil {
		return nil, err
	}

	res, err := s.fromCommonDTO(commonServer)
	if err != nil {
		return nil, err
	}

	res.HideCredentials()

	return res, nil
}

func (s Service) Get(ctx context.Context, id string) (file_server.FileServer, error) {
	dto, err := s.storage.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	res, err := s.fromCommonDTO(dto)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s Service) fromCommonDTO(dto sqlite.CommonFileServerDTO) (file_server.FileServer, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	switch dto.Type {
	case "ssh":
		var res file_server.SSHFileServer
		err := json.Unmarshal([]byte(dto.Params), &res)
		if err != nil {
			return nil, err
		}

		res.ID = dto.ID
		res.Name = dto.Name
		res.Created = dto.Created
		res.Modified = dto.Modified

		return &res, nil
	case "api":
		var res file_server.APIFileServer
		err := json.Unmarshal([]byte(dto.Params), &res)
		if err != nil {
			return nil, err
		}

		res.ID = dto.ID
		res.Name = dto.Name
		res.Created = dto.Created
		res.Modified = dto.Modified

		return &res, nil
	default:
		return nil, errors.New("unknown file server type")
	}
}

func (s Service) StoreChunk(ctx context.Context, fileServer file_server.FileServer, file Opener, start, size int64) (string, error) {
	var (
		res string
		err error
	)

	switch fs := fileServer.(type) {
	case *file_server.SSHFileServer:
		res, err = s.storeOnSSH(ctx, fs, file, start, size)
	case *file_server.APIFileServer:
		res, err = s.storeOnAPI(ctx, fs, file, start, size)
	default:
		return "", errors.New("unknown file server type")
	}

	if err != nil {
		return "", err
	}

	return res, nil
}

func (s Service) Count(ctx context.Context) (int, error) {
	res, err := s.storage.Count(ctx)
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (s Service) storeOnSSH(ctx context.Context, fs *file_server.SSHFileServer, file Opener, start, size int64) (string, error) {
	client, closeFunc, err := ssh.NewClient(fs.Address, fs.Port, fs.User, fs.Key)
	if err != nil {
		return "", err
	}
	defer closeFunc()

	dir := buildFilePath()

	fileName, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	relativePath := path.Join(dir, fileName.String())

	dst := path.Join(fs.BasePath, relativePath)

	err = client.MkdirAll(path.Join(fs.BasePath, dir))
	if err != nil {
		return "", err
	}

	dstFile, err := client.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	f, err := file.Open()
	if err != nil {
		return "", err
	}
	_, err = f.Seek(start, io.SeekStart)
	if err != nil {
		return "", err
	}

	_, err = io.CopyN(dstFile, f, size)
	if err != nil {
		return "", err
	}

	return relativePath, nil
}

func (s Service) storeOnAPI(ctx context.Context, fs *file_server.APIFileServer, file Opener, start, size int64) (string, error) {
	return "", nil
}

func buildFilePath() string {
	now := time.Now()

	intTemp := []int{now.Year(), int(now.Month()), now.Day(), now.Hour()}

	strTemp := make([]string, len(intTemp))
	for i, value := range intTemp {
		strTemp[i] = strconv.Itoa(value)
	}

	return path.Join(strTemp...)
}

func (s Service) CopyChunkToLocal(ctx context.Context, chnk chunk.Chunk, writePos int64, localFileName string) error {
	commonFileServer, err := s.storage.Get(ctx, chnk.FileServerID)
	if err != nil {
		return err
	}

	fileServer, err := s.fromCommonDTO(commonFileServer)
	if err != nil {
		return err
	}

	switch fs := fileServer.(type) {
	case *file_server.SSHFileServer:
		err = s.copyFromSSH(ctx, fs, chnk, writePos, localFileName)
	case *file_server.APIFileServer:
		err = s.copyFromAPI(ctx, fs, chnk, writePos, localFileName)
	default:
		return errors.New("unknown file server type")
	}

	if err != nil {
		return err
	}

	return nil
}

func (s Service) copyFromSSH(ctx context.Context, fileServer *file_server.SSHFileServer, chnk chunk.Chunk, writePos int64, localFileName string) error {
	client, closeFunc, err := ssh.NewClient(fileServer.Address, fileServer.Port, fileServer.User, fileServer.Key)
	if err != nil {
		return err
	}
	defer closeFunc()

	remoteFile, err := client.Open(path.Join(fileServer.BasePath, chnk.FilePath))
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	localFile, err := os.OpenFile(localFileName, os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer func() {
		err = localFile.Close()
		if err != nil {
			s.l.Error(err)
			return
		}
	}()

	_, err = localFile.Seek(writePos, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) copyFromAPI(ctx context.Context, fileServer *file_server.APIFileServer, chnk chunk.Chunk, writePos int64, localFileName string) error {
	return nil
}

func (s Service) OpenChunkFile(ctx context.Context, chnk chunk.Chunk) (io.ReadSeeker, error) {
	commonFileServer, err := s.storage.Get(ctx, chnk.FileServerID)
	if err != nil {
		return nil, err
	}

	fileServer, err := s.fromCommonDTO(commonFileServer)
	if err != nil {
		return nil, err
	}

	var res io.ReadSeeker

	switch fs := fileServer.(type) {
	case *file_server.SSHFileServer:
		res, err = s.openOnSSH(ctx, fs, chnk)
	case *file_server.APIFileServer:
		res, err = s.openOnAPI(ctx, fs, chnk)
	default:
		return nil, errors.New("unknown file server type")
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s Service) openOnSSH(ctx context.Context, fileServer *file_server.SSHFileServer, chnk chunk.Chunk) (io.ReadSeeker, error) {
	// TODO close connection and file

	client, _, err := ssh.NewClient(fileServer.Address, fileServer.Port, fileServer.User, fileServer.Key)
	if err != nil {
		return nil, err
	}

	remoteFile, err := client.Open(path.Join(fileServer.BasePath, chnk.FilePath))
	if err != nil {
		return nil, err
	}

	return remoteFile, nil
}

func (s Service) openOnAPI(ctx context.Context, fileServer *file_server.APIFileServer, chnk chunk.Chunk) (io.ReadSeeker, error) {
	return nil, nil
}
