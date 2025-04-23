package server

import (
	"io"
	"strings"

	"github.com/rs/zerolog/log"
)

type logAdapter struct {
	io.Writer
}

// ConnectionLogFilter defines a function type that takes a string as input
// and returns a boolean. It is used to filter connection logs based on
// specific criteria, where the input string typically represents a log entry
// or a related identifier.
type ConnectionLogFilter func(string) bool

// filters is a slice of logFilter functions used to filter out specific log messages
// related to TLS handshake errors. These filters are designed to ignore known,
// non-critical errors that occur during TLS handshakes, such as aborted connections
// or resets, which are expected behavior in certain scenarios.
var filters = []ConnectionLogFilter{
	func(s string) bool {
		// First filter: Swallow specific TLS handshake errors
		// On the Go issue tracker, users repeatedly report that random scanners, health‑checks or browsers will open
		// a TCP socket to port 443 and then immediately close it (often sending exactly zero TLS bytes),
		// producing an EOF. Those issues were closed as “expected behavior” as they’re not
		// crashes or cert bugs, just aborted handshakes
		// Start here: https://github.com/golang/go/issues/56382
		return strings.Contains(s, "http: TLS handshake error") && strings.HasSuffix(s, "EOF\n")
	},
	func(s string) bool {
		// This filter ignores log messages containing "http: TLS handshake error"
		// with "read: connection reset by peer". This is to handle cases where the connection
		// is reset by the client during the handshake process.
		return strings.Contains(s, "http: TLS handshake error") && strings.Contains(s, "read: connection reset by peer")
	},
}

// AddFilter allows adding a new logFilter to the filters slice.
// This can be used to dynamically add filters for log messages to ignore.
func AddFilter(filter ConnectionLogFilter) {
	filters = append(filters, filter)
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
