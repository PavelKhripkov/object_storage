package file_server_service

import (
	"context"
	"encoding/json"
	"github.com/PavelKhripkov/object_storage/internal/adapter/db/sqlite"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/chunk_model"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/file_server_model"
	"github.com/PavelKhripkov/object_storage/pkg/client/ssh"
	"github.com/gofrs/uuid/v5"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
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

func (s Service) ChooseOneExcluding(ctx context.Context, exclude map[string]bool) (file_server_model.FileServer, error) {
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

func (s Service) Ping(ctx context.Context, fileServer file_server_model.FileServer) error {
	switch fs := fileServer.(type) {
	case *file_server_model.SSHFileServer:
		_, closeFunc, err := ssh.NewClient(ctx, fs.Host, fs.Port, fs.User, fs.Key)
		if err != nil {
			return err
		}
		if err := closeFunc(); err != nil {
			s.l.Error(err)
		}
	case *file_server_model.APIFileServer:
		return nil

	}

	return nil
}

func (s Service) Add(ctx context.Context, dto AddFileServerDTO) (file_server_model.FileServer, error) {
	now := time.Now()
	newID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	commonServer := sqlite.CommonFileServerDTO{
		ID:         newID.String(),
		Name:       dto.GetName(),
		Type:       dto.GetType(),
		Status:     file_server_model.FileServerStatusUnknown,
		TotalSpace: dto.GetTotalSpace(),
		Created:    now,
		Modified:   now,
	}

	params, err := dto.MarshalParams()
	if err != nil {
		return nil, err
	}

	commonServer.Params = params

	if err = s.storage.Add(ctx, commonServer); err != nil {
		return nil, err
	}

	res, err := s.fromCommonDTO(commonServer)
	if err != nil {
		return nil, err
	}

	// TODO come up with better decision.
	resToPing, err := s.fromCommonDTO(commonServer)
	if err != nil {
		return nil, err
	}

	go func() {
		pingCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		status := file_server_model.FileServerStatusOK
		if err := s.Ping(pingCtx, resToPing); err != nil {
			s.l.Error(err)
			status = file_server_model.FileServerStatusFail
		}

		err := s.UpdateStatus(pingCtx, resToPing.GetID(), status)
		if err != nil {
			s.l.Error(err)
		}
	}()

	res.HideCredentials()

	return res, nil
}

func (s Service) UpdateStatus(ctx context.Context, id string, status file_server_model.Status) error {
	return s.storage.UpdateStatus(ctx, id, status)
}

func (s Service) Get(ctx context.Context, id string) (file_server_model.FileServer, error) {
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

func (s Service) fromCommonDTO(dto sqlite.CommonFileServerDTO) (file_server_model.FileServer, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	switch dto.Type {
	case "ssh":
		var res file_server_model.SSHFileServer
		err := json.Unmarshal([]byte(dto.Params), &res)
		if err != nil {
			return nil, err
		}

		res.ID = dto.ID
		res.Name = dto.Name
		res.Status = dto.Status
		res.TotalSpace = dto.TotalSpace
		res.UsedSpace = dto.UsedSpace
		res.Created = dto.Created
		res.Modified = dto.Modified

		return &res, nil
	case "api":
		var res file_server_model.APIFileServer
		err := json.Unmarshal([]byte(dto.Params), &res)
		if err != nil {
			return nil, err
		}

		res.ID = dto.ID
		res.Name = dto.Name
		res.Status = dto.Status
		res.TotalSpace = dto.TotalSpace
		res.UsedSpace = dto.UsedSpace
		res.Created = dto.Created
		res.Modified = dto.Modified

		return &res, nil
	default:
		return nil, errors.New("unknown file server type")
	}
}

func (s Service) Count(ctx context.Context) (int, error) {
	res, err := s.storage.Count(ctx)
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (s Service) UpdateUsedSpace(ctx context.Context, id string, change int64) error {
	return s.storage.UpdateUsedSpace(ctx, id, change)
}

func (s Service) StoreChunk(ctx context.Context, fileServer file_server_model.FileServer, file *multipart.FileHeader, start, size int64) (string, error) {
	var (
		res string
		err error
	)

	switch fs := fileServer.(type) {
	case *file_server_model.SSHFileServer:
		res, err = s.storeOnSSH(ctx, fs, file, start, size)
	case *file_server_model.APIFileServer:
		res, err = s.storeOnAPI(ctx, fs, file, start, size)
	default:
		return "", errors.New("unknown file server type")
	}

	if err != nil {
		return "", err
	}

	return res, nil
}

func (s Service) storeOnSSH(ctx context.Context, fs *file_server_model.SSHFileServer, file *multipart.FileHeader, start, size int64) (string, error) {
	client, closeFunc, err := ssh.NewClient(ctx, fs.Host, fs.Port, fs.User, fs.Key)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := closeFunc(); err != nil {
			s.l.Error(err)
		}
	}()

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
	defer func() {
		if err := dstFile.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	f, err := file.Open()
	if err != nil {
		return "", err
	}

	defer func() {
		if err := f.Close(); err != nil {
			s.l.Error(err)
		}
	}()

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

func (s Service) storeOnAPI(ctx context.Context, fs *file_server_model.APIFileServer, file *multipart.FileHeader, start, size int64) (string, error) {
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

func (s Service) OpenChunkFile(ctx context.Context, chnk chunk_model.Chunk) (func() (io.ReadSeekCloser, error), error) {
	commonFileServer, err := s.storage.Get(ctx, chnk.FileServerID)
	if err != nil {
		return nil, err
	}

	fileServer, err := s.fromCommonDTO(commonFileServer)
	if err != nil {
		return nil, err
	}

	var res func() (io.ReadSeekCloser, error)

	switch fs := fileServer.(type) {
	case *file_server_model.SSHFileServer:
		res, err = s.openOnSSH(ctx, fs, chnk)
	case *file_server_model.APIFileServer:
		res, err = s.openOnAPI(ctx, fs, chnk)
	default:
		return nil, errors.New("unknown file server type")
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s Service) openOnSSH(ctx context.Context, fileServer *file_server_model.SSHFileServer, chnk chunk_model.Chunk) (func() (io.ReadSeekCloser, error), error) {

	res := func() (io.ReadSeekCloser, error) {
		client, closeFunc, err := ssh.NewClient(ctx, fileServer.Host, fileServer.Port, fileServer.User, fileServer.Key)
		if err != nil {
			return nil, err
		}

		remoteFile, err := client.Open(path.Join(fileServer.BasePath, chnk.FilePath))
		if err != nil {
			return nil, err
		}

		res := &sftpWrapper{
			closeClient: closeFunc,
			file:        remoteFile,
		}

		return res, nil
	}

	return res, nil
}

func (s Service) openOnAPI(ctx context.Context, fileServer *file_server_model.APIFileServer, chnk chunk_model.Chunk) (func() (io.ReadSeekCloser, error), error) {
	return nil, nil
}

type sftpWrapper struct {
	closeClient func() error
	file        *sftp.File
}

func (s *sftpWrapper) Read(p []byte) (n int, err error) {
	return s.file.Read(p)
}

func (s *sftpWrapper) Seek(offset int64, whence int) (int64, error) {
	return s.file.Seek(offset, whence)
}

func (s *sftpWrapper) Close() error {
	err := s.file.Close()
	if err != nil {
		return err
	}

	err = s.closeClient()
	if err != nil {
		return err
	}

	return nil
}
