package core

import (
	"fmt"
	"sort"
	"strings"
)

// FullyConnected checks that every leaf has an upstream source. Builtin outputs
// are produced by their operator function, and root inputs are supplied by the
// caller; neither needs an incoming connection. This checks wiring, not whether
// an operator will emit a value for a particular input.
func (p *Port) FullyConnected() error {
	var missing []string
	p.WalkPrimitivePorts(func(leaf *Port) {
		seen := make(map[*Port]bool)
		for source := leaf; source != nil && !seen[source]; source = source.src {
			seen[source] = true
			if source.operator != nil {
				if source.direction == DIRECTION_OUT && source.operator.Builtin() {
					return
				}
				if source.direction == DIRECTION_IN && source.operator.Parent() == nil {
					return
				}
			}
		}
		missing = append(missing, leaf.Name())
	})
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return fmt.Errorf("unconnected output ports: %s", strings.Join(missing, ", "))
}
