package achievements

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"

	pevents "github.com/goverland-labs/goverland-platform-events/events/inbox"
	client "github.com/goverland-labs/goverland-platform-events/pkg/natsclient"

	"github.com/goverland-labs/goverland-inbox-storage/internal/config"
)

const (
	groupName                = "system"
	maxPendingAckPerConsumer = 10
)

type closable interface {
	Close() error
}

type Consumer struct {
	conn      *nats.Conn
	service   *Service
	consumers []closable
}

func NewConsumer(nc *nats.Conn, s *Service) (*Consumer, error) {
	c := &Consumer{
		conn:      nc,
		service:   s,
		consumers: make([]closable, 0),
	}

	return c, nil
}

func (c *Consumer) handler() pevents.AchievementRecalculateHandler {
	return func(payload pevents.AchievementRecalculateEvent) error {
		atype, err := convertAchievementType(payload.Type)
		if err != nil {
			log.Warn().Err(err).Msg("unknown achievement type")

			return nil
		}

		if err := c.service.recalc(context.TODO(), payload.UserID, atype); err != nil {
			log.Error().Err(err).Msg("process event")

			return err
		}

		return nil
	}
}

func (c *Consumer) init() pevents.AchievementInitHandler {
	return func(payload pevents.AchievementInitEvent) error {
		if err := c.service.init(context.TODO(), payload.UserID); err != nil {
			log.Error().Err(err).Msg("process event")

			return err
		}

		return nil
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	group := config.GenerateGroupName(groupName)
	ic, err := client.NewConsumer(ctx, c.conn, group, pevents.SubjectInitAchievement, c.init(), client.WithMaxAckPending(maxPendingAckPerConsumer))
	if err != nil {
		return fmt.Errorf("consume for %s/%s: %w", group, pevents.SubjectInitAchievement, err)
	}

	sc, err := client.NewConsumer(ctx, c.conn, group, pevents.SubjectRecalculateAchievement, c.handler(), client.WithMaxAckPending(maxPendingAckPerConsumer))
	if err != nil {
		return fmt.Errorf("consume for %s/%s: %w", group, pevents.SubjectRecalculateAchievement, err)
	}

	c.consumers = append(c.consumers, ic, sc)

	log.Info().Msg("achievement consumers are started")

	// todo: handle correct stopping the consumer by context
	<-ctx.Done()
	return c.stop()
}

func (c *Consumer) stop() error {
	for _, cs := range c.consumers {
		if err := cs.Close(); err != nil {
			log.Error().Err(err).Msg("close achievement consumer")
		}
	}

	return nil
}
