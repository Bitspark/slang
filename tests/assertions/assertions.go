package assertions

import (
	"slang/core"

	"github.com/stretchr/testify/assert"
)

type SlAssertions struct {
	*assert.Assertions
}

func New(t assert.TestingT) *SlAssertions {
	return &SlAssertions{assert.New(t)}
}

func (sla *SlAssertions) PortPushes(exp []interface{}, p *core.Port) {
	for _, e := range exp {
		a := p.Pull()
		sla.Equal(e, a)
	}
}
