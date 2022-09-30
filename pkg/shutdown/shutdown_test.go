package shutdown_test

import (
	"errors"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/qulaz/artforintrovert-test/pkg/shutdown"
)

func Example() {
	h := http.NewServeMux()
	h.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})

	server := &http.Server{ //nolint:exhaustruct
		Addr:         ":8000",
		Handler:      h,
		ReadTimeout:  time.Second * time.Duration(5),
		WriteTimeout: time.Second * time.Duration(7),
	}

	go shutdown.Graceful(
		[]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM},
		server, // add io.Closer items here
	)

	if err := server.ListenAndServe(); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			log.Println("Server shutdown successfully")
		default:
			log.Fatal(err)
		}
	}
}
