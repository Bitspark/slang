# Port and operator shutdown

Ports carry data, including legitimate `nil` values and stream markers. Shutdown
is separate from that data. `Receive()` returns `(value, nil)` for a complete
value and `(nil, ErrPortClosed)` after closure. Buffered values can be drained;
an incomplete map or stream is discarded when one of its reads is cancelled.
`Poll()` retains its existing signature and returns `(nil, false)` for either
timeout or closure.

`Pull()` and its typed helpers keep their single-value operator API. On closure
they panic with `ErrPortClosed`, which unwinds the current operator invocation
and runs its defers. The operator execution boundary handles that specific
cancellation without logging a failure. This prevents existing operator loops
from processing an invented zero value or spinning on an unfinished stream.
External consumers should use `Receive()`. Custom operator workers must use
`Operator.Go`; callbacks owned by another library can use `Operator.Run`.

`Stop()` broadcasts cancellation to the operator hierarchy and closes its input
and output ports. It is safe to call concurrently. `Done()` and `WaitForStop()`
signal cancellation; they do not imply that every worker has returned. A
subsequent `Start()` joins the old workers across the entire hierarchy before
reopening ports. Configure and connect graphs before starting them, and call
`Start()` only after `Stop()` has returned. Concurrent start/reconfiguration is
not supported. External callbacks and I/O must also respond to cancellation.

`Fail(err)` stops the graph and retains its first failure, available through
`Err()` on any operator in the hierarchy. Unexpected panics are logged and
recorded through that path. An ordinary `Stop()` leaves `Err()` nil. This gives
hosts a failure cause; it does not introduce an error value into Slang streams
or define graph-level recovery/retry operators.

Each opening owns a separate buffer. All sends, receives, resizes, and closure
use its mutex. A full-buffer sender releases that mutex before waiting on a
change notification, allowing close to wake it without a send/close race or
deadlock. Waiters keep the old buffer when a port reopens. The store protects
its own data instead of exposing a port's internal lock.

Regression coverage includes concurrent send/close, blocked writers/readers,
valid nil data, partial streams, dynamic growth, reopening, operator restart,
failure propagation, and HTTP delegate cancellation. CI runs the complete Go
suite with `-race -count=3` on Linux, alongside ordinary platform tests.
