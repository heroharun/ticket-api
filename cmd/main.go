package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	http "ticket-purchase/cmd/server"
	"ticket-purchase/internal/api"
	"ticket-purchase/internal/store"
	"ticket-purchase/pkg/postgres"
	"ticket-purchase/pkg/sysignals"
	"time"
)

var exitErr error

func main() {
	ctx := context.Background()

	con := postgres.CreateConnection()
	err := postgres.ManuelMigrator(con)
	if err != nil {
		fmt.Errorf("error migrating: %v", err)
	}
	store := store.New(con)
	api := api.New(store)

	httpConfig := http.Config{
		Host:              "",
		Port:              8080,
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	httpServer := http.New(&httpConfig, api, true)
	defer shutdown(
		ctx,
		httpServer,
		con,
	)
	fatalErr := make(chan error, 1)
	go sysignals.NotifyErrorOnQuit(fatalErr)

	startServices(
		fatalErr,
		httpConfig,
		httpServer,
	)

	go liveHealthcheck(
		httpServer,
		time.Second*30,
		"purchase-ticket@v0.0.1",
	)
	exitErr = <-fatalErr
}

func liveHealthcheck(
	httpSvc *http.HTTP,
	delay time.Duration,
	appVersion string,
) {
	httpSvc.AppendHealthResponse("version", appVersion)
	for {
		time.Sleep(delay)
	}
}

func startServices(fatalErr chan error, config http.Config, httpServer *http.HTTP) {
	// HTTP server is anyway started even if there are no APIs, for healthcheck
	go func() {
		httpConfig := config
		fmt.Println("[http] listening on %s:%d", httpConfig.Host, httpConfig.Port)
		fatalErr <- httpServer.Start()
	}()
}

func shutdown(ctx context.Context, httpServer *http.HTTP, con *sql.DB) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	wgroup := &sync.WaitGroup{}

	if httpServer != nil {
		wgroup.Add(1)
		go func() {
			defer wgroup.Done()
			err := httpServer.Shutdown(ctx)
			if err != nil {
				fmt.Errorf("http server shutdown failed: %v", err)
			}
		}()
	}
	// wait for the services to shutdown before closing drivers
	wgroup.Wait()

	if con != nil {
		wgroup.Add(1)
		go func() {
			defer wgroup.Done()
			err := con.Close()
			if err != nil {
				fmt.Errorf("database connection close failed: %v", err)
			}
		}()
	}

	wgroup.Wait()
}
