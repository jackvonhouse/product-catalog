package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type Shutdownify interface {
	Shutdown(context.Context) error
}

func Graceful(
	ctx context.Context,
	cancel context.CancelFunc,
	log log.Logger,
	shutdownItems ...Shutdownify,
) {

	defer cancel()

	gCtx, cancelGraceful := signal.NotifyContext(
		ctx,
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT,
		os.Interrupt,
	)

	defer cancelGraceful()

	log.Info("start listening term signals")

	<-gCtx.Done()

	log.Info("term signal received. start closing items")

	for _, shutdownItem := range shutdownItems {
		if err := shutdownItem.Shutdown(ctx); err != nil {
			log.Errorf("error on shutdown item: %s", err.Error())
		}
	}
}
