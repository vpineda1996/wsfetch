package authenticator

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/vpineda1996/wsfetch/pkg/auth/types"
	"go.uber.org/zap"
)

type persistentClient struct {
	delegate Client
}

var (
	_ Client = &persistentClient{}
)

const (
	configFilePath = "authclient.json"
)

// Persistent client is not multitenant but it perserves
// sessions and OTP validations across executions
func NewPersistentClient() (Client, error) {
	c, err := restoreAuthClient(configFilePath)
	if err != nil {
		log.Infow("Could not read client info, creating new one", zap.Error(err))
		c = NewClient()
	}
	return &persistentClient{
		delegate: c,
	}, nil

}

// Authenticate tries to get the cached client in fs
func (p *persistentClient) Authenticate(ctx context.Context, creds types.PasswordCredentials) (*types.Session, error) {
	res, err := p.delegate.Authenticate(ctx, creds)
	if err != nil {
		return nil, err
	}
	saveAuthClient(configFilePath, p.delegate)
	return res, nil
}

func saveAuthClient(filePath string, client Client) {
	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		log.Errorw("Could not create file to save client", zap.Error(err))
		return
	}
	bits, err := json.Marshal(client)
	if err != nil {
		log.Errorw("Could not serialize client")
		return
	}
	n, err := f.Write(bits)
	if err != nil {
		log.Errorw("Unable to save data to file")
	}
	log.Infow("Finished writing file", "bytesWritten", n, "expectedBytesWritten", len(bits))
}

func restoreAuthClient(filePath string) (Client, error) {
	f, err := os.Open(filePath)
	if err != nil {

		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var aClient client
	err = json.Unmarshal(data, &aClient)
	if err != nil {
		return nil, err
	}
	newFromExisting(&aClient)
	return &aClient, nil
}
