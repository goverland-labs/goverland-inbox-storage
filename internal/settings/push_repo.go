package settings

import (
	"errors"
	"fmt"
	"sync"

	vaultapi "github.com/hashicorp/vault/api"
)

const (
	keyData  = "data"
	keyToken = "token"
)

type vaultReadWriter interface {
	Read(path string) (*vaultapi.Secret, error)
	Write(path string, data map[string]interface{}) (*vaultapi.Secret, error)
	Delete(path string) (*vaultapi.Secret, error)
}

var (
	ErrUnableToCastData = errors.New("failed to cast data")
	ErrTokenNotFound    = errors.New("token not found")
)

// PushRepo provides capability to store and get keys in Vault.
// Also this struct has in-memory cache to reduce database queries
type PushRepo struct {
	cli      vaultReadWriter
	basePath string

	cache map[string]string
	mux   sync.Mutex
}

func NewPushRepo(cli vaultReadWriter, path string) *PushRepo {
	return &PushRepo{
		cli:      cli,
		basePath: path,
		cache:    make(map[string]string),
	}
}

func (s *PushRepo) getPath(userID string) string {
	return fmt.Sprintf("%s%s", s.basePath, userID)
}

func (s *PushRepo) GetByUserID(userID string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if token, ok := s.cache[userID]; ok {
		return token, nil
	}

	sec, err := s.cli.Read(s.getPath(userID))
	if err != nil {
		return "", err
	}

	if sec == nil {
		return "", ErrTokenNotFound
	}

	data, ok := sec.Data[keyData].(map[string]interface{})
	if !ok {
		return "", ErrUnableToCastData
	}

	token, ok := data[keyToken].(string)
	if !ok {
		return "", ErrUnableToCastData
	}

	s.cache[userID] = token

	return token, nil
}

func (s *PushRepo) Save(userID, token string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	_, err := s.cli.Write(s.getPath(userID), map[string]interface{}{
		keyData: map[string]interface{}{
			keyToken: token,
		},
	})
	if err != nil {
		return err
	}

	s.cache[userID] = token

	return nil
}

func (s *PushRepo) Delete(userID string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	_, err := s.cli.Delete(s.getPath(userID))
	if err != nil {
		return err
	}

	delete(s.cache, userID)

	return nil
}
