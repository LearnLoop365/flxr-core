package servers

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func HandleGracefulShutdown(server *http.Server, appName string, cleanup func(), timeout time.Duration) {
	// Bidirectional Chan for os.Signal values
	// buffer size = 1 -> so RECEIVE OP doesn't block SENDER when no value is sent
	osSignalChannel := make(chan os.Signal, 1)

	// Registers the channel with the Go runtime's signal delivery system
	// signal.Nofity sends one of the listed os signals only when the OS delivers the signal
	signal.Notify(osSignalChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// Wait for a signal
		sig := <-osSignalChannel
		log.Printf("[INFO] Got signal: `%s`. Terminating `%s`...", sig, appName)

		// Cleanup resources. e.g. prepared statements, db connections, etc
		if cleanup != nil {
			cleanup()
		}

		// Context with timeout for server shutdown
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Server Shutdown to Stop Accepting New HTTP Requests Immediately
		// But with the context with timeout, requests already being processed get time to finish
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("[ERROR] failed to close the Go http server: %v", err)
		}
		log.Println("[INFO] Shutdown complete. Exiting.")
	}()
}
