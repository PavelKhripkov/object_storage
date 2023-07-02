package file_server

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/file_server"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/item_data"
	"os"

	"net/url"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SSHFileService struct {
	id string
}

func (s SSHFileService) ID() string {
	return s.id
}

func (s SSHFileService) Status(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (s SSHFileService) List(ctx context.Context) ([]file_server.SSHFileServer, error) {
	return nil, nil
}

func (s SSHFileService) ChooseOne(ctx context.Context, exclude map[string]file_server.SSHFileServer) (*file_server.SSHFileServer, error) {
	return nil, nil
}

type AddParams struct {
	url url.URL
}

func (s SSHFileService) Store(ctx context.Context, file File, start, limit int64) error {

	return nil
}

func (s SSHFileService) GetPart(ctx context.Context, part item_data.Part) ([]byte, error) {
	return nil, nil
}

func SSHCopyFile(srcPath, dstPath string) error {
	config := &ssh.ClientConfig{
		User: "user",
		Auth: []ssh.AuthMethod{
			ssh.Password("pass"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, _ := ssh.Dial("tcp", "remotehost:22", config)
	defer client.Close()

	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftp.Close()

	// Open the source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := sftp.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// write to file
	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return err
	}
	return nil
}
