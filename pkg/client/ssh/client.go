package ssh

import (
	"context"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"time"
)

func NewClient(ctx context.Context, host, port, user, key string) (*sftp.Client, func() error, error) {
	signer, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		return nil, nil, err
	}

	var timeOut time.Duration

	if deadLine, ok := ctx.Deadline(); ok {
		timeOut = deadLine.Sub(time.Now())
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         timeOut,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO remove in prod
	}

	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		return nil, nil, err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, nil, err
	}

	closeFunc := func() error {
		err := sftpClient.Close()
		if err != nil {
			return err
		}

		err = client.Close()
		if err != nil {
			return err
		}
		return nil
	}

	return sftpClient, closeFunc, nil
}
