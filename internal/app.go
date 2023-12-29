package internal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	coresdk "github.com/goverland-labs/core-web-sdk"
	"github.com/goverland-labs/helpers-ens-resolver/proto"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/goverland-labs/platform-events/pkg/natsclient"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/nats-io/nats.go"
	"github.com/s-larionov/process-manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/goverland-labs/inbox-storage/internal/config"
	"github.com/goverland-labs/inbox-storage/internal/settings"
	"github.com/goverland-labs/inbox-storage/internal/subscription"
	"github.com/goverland-labs/inbox-storage/internal/user"
	"github.com/goverland-labs/inbox-storage/pkg/grpcsrv"
	"github.com/goverland-labs/inbox-storage/pkg/health"
	"github.com/goverland-labs/inbox-storage/pkg/prometheus"
)

type Application struct {
	sigChan <-chan os.Signal
	manager *process.Manager
	cfg     config.App
	db      *gorm.DB

	us        *user.Service
	sub       *subscription.Service
	settings  *settings.Service
	ensClient proto.EnsClient
}

func NewApplication(cfg config.App) (*Application, error) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	a := &Application{
		sigChan: sigChan,
		cfg:     cfg,
		manager: process.NewManager(),
	}

	err := a.bootstrap()
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Application) Run() {
	a.manager.StartAll()
	a.registerShutdown()
}

func (a *Application) bootstrap() error {
	initializers := []func() error{
		a.initDB,
		a.initEnsResolver,
		// Init Dependencies
		a.initServices,

		// Init Workers: Application
		a.initAPI,

		// Init Workers: System
		a.initPrometheusWorker,
		a.initHealthWorker,
	}

	for _, initializer := range initializers {
		if err := initializer(); err != nil {
			return err
		}
	}

	return nil
}

func (a *Application) initDB() error {
	db, err := gorm.Open(postgres.Open(a.cfg.DB.DSN), &gorm.Config{})
	if err != nil {
		return err
	}

	ps, err := db.DB()
	if err != nil {
		return err
	}
	ps.SetMaxOpenConns(a.cfg.DB.MaxOpenConnections)

	a.db = db
	if a.cfg.DB.Debug {
		a.db = db.Debug()
	}

	return err
}

func (a *Application) initEnsResolver() error {
	conn, err := grpc.Dial(a.cfg.API.EnsResolverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create connection with ens resolver: %v", err)
	}

	a.ensClient = proto.NewEnsClient(conn)

	return nil
}

func (a *Application) initServices() error {
	nc, err := nats.Connect(
		a.cfg.Nats.URL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(a.cfg.Nats.MaxReconnects),
		nats.ReconnectWait(a.cfg.Nats.ReconnectTimeout),
	)
	if err != nil {
		return err
	}

	pb, err := natsclient.NewPublisher(nc)
	if err != nil {
		return err
	}

	_ = pb

	a.initUsers()
	if err = a.initSubscription(); err != nil {
		return err
	}
	if err = a.initPushes(); err != nil {
		return err
	}

	return nil
}

func (a *Application) initPushes() error {
	vc, err := vaultapi.NewClient(&vaultapi.Config{
		Address: a.cfg.Vault.Address,
	})
	if err != nil {
		return err
	}

	vc.SetToken(a.cfg.Vault.Token)
	repo := settings.NewPushRepo(vc.Logical(), a.cfg.Vault.BasePath)
	service := settings.NewService(repo)

	a.settings = service

	return nil
}

func (a *Application) initUsers() {
	repo := user.NewRepo(a.db)
	sessionRepo := user.NewSessionRepo(a.db)
	a.us = user.NewService(repo, sessionRepo, a.ensClient)

	ensWorker := user.NewEnsResolverWorker(repo, a.ensClient)
	a.manager.AddWorker(process.NewCallbackWorker("ens_resolver", ensWorker.Start))
}

func (a *Application) initPrometheusWorker() error {
	srv := prometheus.NewServer(a.cfg.Prometheus.Listen, "/metrics")
	a.manager.AddWorker(process.NewServerWorker("prometheus", srv))

	return nil
}

func (a *Application) initHealthWorker() error {
	srv := health.NewHealthCheckServer(a.cfg.Health.Listen, "/status", health.DefaultHandler(a.manager))
	a.manager.AddWorker(process.NewServerWorker("health", srv))

	return nil
}

func (a *Application) initSubscription() error {
	repo := subscription.NewRepo(a.db)
	globalRepo := subscription.NewGlobalRepo(a.db)
	cache := subscription.NewCache()
	cs := coresdk.NewClient(a.cfg.Core.CoreURL)

	feedConn, err := grpc.Dial(a.cfg.API.FeedAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create connection with storage server: %v", err)
	}
	fc := inboxapi.NewFeedClient(feedConn)
	service, err := subscription.NewService(repo, globalRepo, cache, a.cfg.Core.SubscriberID, cs, fc)
	if err != nil {
		return fmt.Errorf("subscription service: %w", err)
	}

	a.sub = service

	return nil
}

// todo: move exclude path to config?
func (a *Application) initAPI() error {
	authInterceptor := grpcsrv.NewAuthInterceptor()
	srv := grpcsrv.NewGrpcServer(
		[]string{
			"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		},
		authInterceptor.AuthAndIdentifyTickerFunc,
	)

	inboxapi.RegisterSubscriptionServer(srv, subscription.NewServer(a.sub))
	inboxapi.RegisterUserServer(srv, user.NewServer(a.us))
	inboxapi.RegisterSettingsServer(srv, settings.NewServer(a.settings, a.us))

	a.manager.AddWorker(grpcsrv.NewGrpcServerWorker("API", srv, a.cfg.API.Bind))

	return nil
}

func (a *Application) registerShutdown() {
	go func(manager *process.Manager) {
		<-a.sigChan

		manager.StopAll()
	}(a.manager)

	a.manager.AwaitAll()
}
