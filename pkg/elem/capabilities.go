package elem

import (
	"net/http"
	"time"
)

// Capabilities are the host resources an operator may use. Operators receive them
// when they are built instead of reaching the operating system themselves, so the
// host decides what a program can touch: the local CLI passes this machine, tests
// pass fakes, and a hosted runner can pass implementations that call a separate
// capability service. A nil capability is not provided, and an operator that needs
// it cannot be built.
type Capabilities struct {
	// HTTP sends the HTTP client operator's requests. Its Timeout bounds a whole
	// exchange, including reading the response body.
	HTTP *http.Client
}

// localHTTPTimeout stays below the hosted runtime's 15-second invocation limit, so
// a program receives a failed request instead of the runtime ending the invocation.
const localHTTPTimeout = 10 * time.Second

// LocalCapabilities reach this machine directly. They reproduce how operators
// behaved before capabilities existed.
func LocalCapabilities() Capabilities {
	return Capabilities{
		HTTP: &http.Client{Timeout: localHTTPTimeout},
	}
}
