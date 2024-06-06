package settings

import (
	"errors"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
)

const (
	keyData  = "data"
	keysData = "keys"
	keyToken = "token"
)

type vaultReadWriter interface {
	List(path string) (*vaultapi.Secret, error)
	Read(path string) (*vaultapi.Secret, error)
	Write(path string, data map[string]interface{}) (*vaultapi.Secret, error)
	Delete(path string) (*vaultapi.Secret, error)
}

type PushDetails struct {
	DeviceUUID string
	Token      string
}

var (
	ErrUnableToCastData = errors.New("failed to cast data")
	ErrTokenNotFound    = errors.New("token not found")
)

// PushRepo provides capability to store and get keys in Vault.
type PushRepo struct {
	cli      vaultReadWriter
	basePath string
}

func NewPushRepo(cli vaultReadWriter, path string) *PushRepo {
	return &PushRepo{
		cli:      cli,
		basePath: strings.TrimRight(path, "/"),
	}
}

func (s *PushRepo) getPathV1(userID string) string {
	return fmt.Sprintf("%s/%s", s.basePath, userID)
}

func (s *PushRepo) getPathByUser(userID string) string {
	return fmt.Sprintf("%s/tokens_by_devices/%s", s.basePath, userID)
}

func (s *PushRepo) getPathByUserDevice(userID, deviceUUID string) string {
	return fmt.Sprintf("%s/%s", s.getPathByUser(userID), deviceUUID)
}

func (s *PushRepo) GetListByUserID(userID string) ([]PushDetails, error) {
	sec, err := s.cli.Read(s.getPathByUser(userID))
	if err != nil {
		return nil, err
	}

	if sec == nil {
		// let's try fallback logic
		token, err := s.GetByUserID(userID)
		if err != nil && !errors.Is(err, ErrTokenNotFound) {
			return nil, fmt.Errorf("fallback by user: %w", err)
		}

		if errors.Is(err, ErrTokenNotFound) {
			return nil, nil
		}

		return []PushDetails{
			{
				DeviceUUID: "default_device",
				Token:      token,
			},
		}, nil
	}

	data, ok := sec.Data[keysData].([]interface{})
	if !ok {
		return nil, ErrUnableToCastData
	}

	result := make([]PushDetails, 0, len(data))
	for idx := range data {
		deviceUUID, ok := data[idx].(string)
		if !ok {
			return nil, fmt.Errorf("cast device uuid: %w", ErrUnableToCastData)
		}

		token, err := s.GetByUserAndDevice(userID, deviceUUID)
		if err != nil {
			return nil, fmt.Errorf("get token by device : %w", err)
		}

		result = append(result, PushDetails{
			DeviceUUID: deviceUUID,
			Token:      token,
		})
	}

	return result, nil
}

// GetByUserID returns saved user token
// @deprecated: use GetByUserAndDevice instead
func (s *PushRepo) GetByUserID(userID string) (string, error) {
	sec, err := s.cli.Read(s.getPathV1(userID))
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

	return token, nil
}

func (s *PushRepo) GetByUserAndDevice(userID, deviceUUID string) (token string, err error) {
	sec, err := s.cli.Read(s.getPathByUserDevice(userID, deviceUUID))
	if err != nil {
		return "", err
	}

	defer func() {
		if err == nil {
			return
		}

		// fallback for getting by old version and resaving by new path
		token, err = s.GetByUserID(userID)
		if err != nil {
			return
		}

		if errSave := s.Save(userID, deviceUUID, token); errSave != nil {
			log.Err(err).Msg("failed to resave token")
		}
	}()

	if sec == nil {
		return "", ErrTokenNotFound
	}

	data, ok := sec.Data[keyData].(map[string]interface{})
	if !ok {
		return "", ErrUnableToCastData
	}

	token, ok = data[keyToken].(string)
	if !ok {
		return "", ErrUnableToCastData
	}

	return token, nil
}

// Save storing provided token for user by device
func (s *PushRepo) Save(userID, deviceUUID, token string) error {
	_, err := s.cli.Write(s.getPathByUserDevice(userID, deviceUUID), map[string]interface{}{
		keyData: map[string]interface{}{
			keyToken: token,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// Delete remove token from storage by user and device
func (s *PushRepo) Delete(userID, deviceUUID string) error {
	_, err := s.cli.Delete(s.getPathByUserDevice(userID, deviceUUID))
	if err != nil {
		return err
	}

	return nil
}
