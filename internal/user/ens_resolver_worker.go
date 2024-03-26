package user

import (
	"context"
	"fmt"
	"time"

	"github.com/goverland-labs/goverland-helpers-ens-resolver/protocol/enspb"
	"github.com/rs/zerolog/log"
)

const syncInterval = 5 * time.Minute

type EnsResolverWorker struct {
	repo *Repo

	ensClient enspb.EnsClient
}

func NewEnsResolverWorker(repo *Repo, ensClient enspb.EnsClient) *EnsResolverWorker {
	return &EnsResolverWorker{repo: repo, ensClient: ensClient}
}

func (e *EnsResolverWorker) Start(ctx context.Context) error {
	for {
		select {
		case <-time.After(syncInterval):
			err := e.sync(ctx)
			if err != nil {
				log.Error().Err(err).Msg("sync ens names")
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (e *EnsResolverWorker) sync(ctx context.Context) error {
	users, err := e.repo.GetRegularWithoutEnsName()
	if err != nil {
		return fmt.Errorf("get users without ens name: %w", err)
	}

	addresses := make([]string, 0, len(users))
	for _, user := range users {
		if user.Address == nil {
			log.Warn().Msgf("regular user #%s has no address", user.ID)

			continue
		}

		addresses = append(addresses, *user.Address)
	}

	if len(addresses) == 0 {
		return nil
	}

	log.Info().Msgf("sync ens names for %d users", len(addresses))

	resp, err := e.ensClient.ResolveDomains(ctx, &enspb.ResolveDomainsRequest{Addresses: addresses})
	if err != nil {
		return fmt.Errorf("resolve domains: %w", err)
	}

	for _, ensResp := range resp.GetAddresses() {
		err := e.repo.UpdateEnsWhereAddress(ensResp.GetAddress(), ensResp.GetEnsName())
		if err != nil {
			log.Error().Err(err).Msgf("update ens name for address #%s", ensResp.GetAddress())
		}
	}

	return nil
}
