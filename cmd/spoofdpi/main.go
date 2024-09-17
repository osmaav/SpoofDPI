package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/osmaav/SpoofDPI/util/log"

	"github.com/osmaav/SpoofDPI/proxy"
	"github.com/osmaav/SpoofDPI/util"
	"github.com/osmaav/SpoofDPI/version"
)

func main() {
	args := util.ParseArgs()
	if args.Version {
		version.PrintVersion()
		os.Exit(0)
	}

	config := util.GetConfig()
	config.Load(args)

	log.InitLogger(config)
	ctx := util.GetCtxWithScope(context.Background(), "MAIN")
	logger := log.GetCtxLogger(ctx)

	pxy := proxy.New(config)

	if !config.Silent {
		util.PrintColoredBanner()
	}

	if config.SystemProxy {
		if err := util.SetOsProxy(uint16(config.Port)); err != nil {
			logger.Fatal().Msgf("error while changing proxy settings: %s", err)
		}
		defer func() {
			if err := util.UnsetOsProxy(); err != nil {
				logger.Fatal().Msgf("error while disabling proxy: %s", err)
			}
		}()
	}

	go pxy.Start(context.Background())

	// Handle signals
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(
		sigs,
		syscall.SIGKILL,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP)

	go func() {
		_ = <-sigs
		done <- true
	}()

	<-done
}
