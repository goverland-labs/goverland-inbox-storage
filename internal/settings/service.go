package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/goverland-labs/goverland-platform-events/events/inbox"
	"github.com/rs/zerolog/log"
	"go.openly.dev/pointy"
	"gorm.io/gorm"
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

type Publisher interface {
	PublishJSON(ctx context.Context, subject string, obj any) error
}

type Service struct {
	tokens    TokenProvider
	details   DetailsManipulator
	publisher Publisher
}

func NewService(t TokenProvider, dm DetailsManipulator, publisher Publisher) *Service {
	return &Service{
		tokens:    t,
		details:   dm,
		publisher: publisher,
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// default logic for getting pushes
		return getPushDefaultSettings(), nil
	}

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
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("get push details: %w", err)
	}

	psd := getPushDefaultSettings()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		details = &Details{
			UserID: userID,
			Type:   DetailsTypePushConfig,
		}
	} else {
		if err = json.Unmarshal(details.Value, psd); err != nil {
			return fmt.Errorf("unmarshal push details: %w", err)
		}
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

func getPushDefaultSettings() *PushSettingsDetails {
	return &PushSettingsDetails{
		NewProposalCreated: pointy.Bool(true),
		QuorumReached:      pointy.Bool(true),
		VoteFinishesSoon:   pointy.Bool(true),
		VoteFinished:       pointy.Bool(true),
	}
}

func (s *Service) GetFeedSettings(userID uuid.UUID) (*FeedSettings, error) {
	details, err := s.details.GetByUserAndType(userID, DetailsTypeFeedConfig)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// default logic for getting config
		return getDefaultFeedSettings(), nil
	}

	if err != nil {
		return nil, fmt.Errorf("get push details: %w", err)
	}

	var fsd FeedSettings
	if err = json.Unmarshal(details.Value, &fsd); err != nil {
		return nil, fmt.Errorf("unmarshal push details: %w", err)
	}

	return &fsd, nil
}

func (s *Service) StoreFeedSettings(userID uuid.UUID, req FeedSettings) error {
	details, err := s.details.GetByUserAndType(userID, DetailsTypeFeedConfig)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("get push details: %w", err)
	}

	fsd := getDefaultFeedSettings()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		details = &Details{
			UserID: userID,
			Type:   DetailsTypeFeedConfig,
		}
	} else {
		if err = json.Unmarshal(details.Value, fsd); err != nil {
			return fmt.Errorf("unmarshal push details: %w", err)
		}
	}

	if req.ArchiveProposalAfterVote != nil {
		fsd.ArchiveProposalAfterVote = req.ArchiveProposalAfterVote
	}

	if req.AutoarchiveAfterDuration != nil {
		fsd.AutoarchiveAfterDuration = req.AutoarchiveAfterDuration
	}

	raw, err := json.Marshal(fsd)
	if err != nil {
		return fmt.Errorf("marshal push details: %w", err)
	}

	details.Value = raw

	if err = s.details.StoreDetails(details); err != nil {
		return fmt.Errorf("store push details: %w", err)
	}

	go func() {
		d, _ := time.ParseDuration(pointy.StringValue(fsd.AutoarchiveAfterDuration, "7d"))
		days := int(d.Hours() / 24)

		if err = s.publisher.PublishJSON(context.TODO(), inbox.SubjectFeedSettingsUpdated, inbox.FeedSettingsPayload{
			SubscriberID:         details.UserID,
			AutoarchiveAfterDays: days,
		}); err != nil {
			log.Err(err).Msg("publish feed settings update")
		}
	}()

	return nil
}

func getDefaultFeedSettings() *FeedSettings {
	return &FeedSettings{
		ArchiveProposalAfterVote: pointy.Bool(false),
		AutoarchiveAfterDuration: pointy.String("7d"),
	}
}
