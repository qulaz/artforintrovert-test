package shutdown

import (
	"io"
	"log"
	"os"
	"os/signal"
)

// Graceful helper for graceful shutdown.
func Graceful(signals []os.Signal, closeItems ...io.Closer) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, signals...)
	sig := <-sigc
	log.Printf("Caught signal %s. Shutting down...", sig)

	for _, closer := range closeItems {
		if err := closer.Close(); err != nil {
			log.Printf("failed to close %v: %v\n", closer, err)
		}
	}
}
