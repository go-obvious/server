package server

import (
	"io"
	"strings"

	"github.com/rs/zerolog/log"
)

type logAdapter struct {
	io.Writer
}

type logFilter func(string) bool

var filters = []logFilter{
	func(s string) bool {
		// First filter: Swallow specific TLS handshake errors
		// On the Go issue tracker, users repeatedly report that random scanners, health‑checks or browsers will open
		// a TCP socket to port 443 and then immediately close it (often sending exactly zero TLS bytes),
		// producing an EOF. Those issues were closed as “expected behavior” as they’re not
		// crashes or cert bugs, just aborted handshakes
		// Start here: https://github.com/golang/go/issues/56382
		return strings.Contains(s, "http: TLS handshake error") && strings.HasSuffix(s, "EOF\n")
	},
	// Add more filters here as needed
}

func (f logAdapter) Write(p []byte) (int, error) {
	s := string(p)

	// Apply filters
	for _, filter := range filters {
		if filter(s) {
			// Swallow the log message
			return len(p), nil
		}
	}

	// Log the message using zerolog instead
	log.Error().Msg(s)

	// Return the length of p and nil to satisfy the Write method contract
	return len(p), nil
}
