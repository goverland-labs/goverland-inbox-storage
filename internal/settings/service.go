package settings

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type DetailsManipulator interface {
	GetByUserAndType(userID uuid.UUID, dt DetailsType) (*Details, error)
	StoreDetails(info *Details) error
}

type TokenProvider interface {
	GetByUserID(userID string) (string, error)
	GetByUserAndDevice(userID, deviceUUID string) (string, error)
	GetListByUserID(userID string) ([]PushDetails, error)
	Save(userID, deviceUUID, token string) error
	Delete(userID, deviceUUID string) error
}

type Service struct {
	tokens  TokenProvider
	details DetailsManipulator
}

func NewService(t TokenProvider, dm DetailsManipulator) *Service {
	return &Service{
		tokens:  t,
		details: dm,
	}
}

func (s *Service) GetByUserAndDevice(userID, deviceUUID string) (string, error) {
	return s.tokens.GetByUserAndDevice(userID, deviceUUID)
}

func (s *Service) DeleteByUserID(userID, deviceUUID string) error {
	return s.tokens.Delete(userID, deviceUUID)
}

func (s *Service) Upsert(userID, deviceUUID, token string) error {
	if err := s.tokens.Save(userID, deviceUUID, token); err != nil {
		return fmt.Errorf("save token: %s: %w", userID, err)
	}

	return nil
}

func (s *Service) GetListByUserID(userID string) ([]PushDetails, error) {
	list, err := s.tokens.GetListByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("get token list: %w", err)
	}

	return list, nil
}

func (s *Service) GetPushDetails(userID uuid.UUID) (*PushSettingsDetails, error) {
	details, err := s.details.GetByUserAndType(userID, DetailsTypePushConfig)
	if err != nil {
		return nil, fmt.Errorf("get push details: %w", err)
	}

	var psd PushSettingsDetails
	if err = json.Unmarshal(details.Value, &psd); err != nil {
		return nil, fmt.Errorf("unmarshal push details: %w", err)
	}

	return &psd, nil
}

func (s *Service) StorePushDetails(userID uuid.UUID, req PushSettingsDetails) error {
	details, err := s.details.GetByUserAndType(userID, DetailsTypePushConfig)
	if err != nil {
		return fmt.Errorf("get push details: %w", err)
	}

	var psd PushSettingsDetails
	if err = json.Unmarshal(details.Value, &psd); err != nil {
		return fmt.Errorf("unmarshal push details: %w", err)
	}

	if req.NewProposalCreated != nil {
		psd.NewProposalCreated = req.NewProposalCreated
	}

	if req.VoteFinishesSoon != nil {
		psd.VoteFinishesSoon = req.VoteFinishesSoon
	}

	if req.VoteFinished != nil {
		psd.VoteFinished = req.VoteFinished
	}

	if req.QuorumReached != nil {
		psd.QuorumReached = req.QuorumReached
	}

	raw, err := json.Marshal(psd)
	if err != nil {
		return fmt.Errorf("marshal push details: %w", err)
	}

	details.Value = raw

	return s.details.StoreDetails(details)
}
