package ssh

import (
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func NewClient(host, port, user, key string) (*sftp.Client, func() error, error) {
	signer, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		return nil, nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
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
