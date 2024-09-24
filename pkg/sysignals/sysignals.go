package sysignals

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bnkamalesh/errors"
)

var Err = errors.New("quit signal received")

// NotifyErrorOnQuit creates an error upon receiving any of the os signal, to quit the app.
// The error is then pushed to the channel
func NotifyErrorOnQuit(errs chan<- error, otherSignals ...syscall.Signal) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt)
	for signalType := range interrupt {
		switch signalType {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGTSTP:
			errs <- errors.Wrapf(Err, "%v", signalType)
			return
		}

		for _, oSignal := range otherSignals {
			if oSignal == signalType {
				errs <- errors.Wrapf(Err, "%v", signalType)
				return
			}
		}
	}
}
