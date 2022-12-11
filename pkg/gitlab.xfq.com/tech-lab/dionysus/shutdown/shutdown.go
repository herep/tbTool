package shutdown

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const DefaultShutdownTimeOut = time.Second * 15

var quit = make(chan os.Signal)
var sysExit = os.Exit

func init() {

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	signal.Ignore(syscall.SIGHUP)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
}

func NotifyAfterFinish(finishChan chan<- struct{}, runFunc func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// todo:: when testing project: Panic occurred in start process: Fail in goroutine after TestCtlCmd has completed
				log.Printf("[error] Panic occurred in start process: %s", r)
			}
			finishChan <- struct{}{}
		}()

		runFunc()
	}()
}

// waiting for sys.Signal or user's finishChan
func WaitingForNotifies(finishChan <-chan struct{}, shutdownFunc func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[error] Panic occurred in shutdown process: %s", r)
			sysExit(3)
		}
	}()

	select {
	case <-quit:
		log.Println("[info] Shuting down ...")
		shutdownFunc()

	case <-finishChan:
	}

	log.Printf("[Dio] Exited.")
}

// Deprecated:: Use this func in test only
func UnsafeChanForTest() chan os.Signal {
	return quit
}
