package assertions

import (
	"fmt"

	"github.com/Bitspark/slang/pkg/core"

	"github.com/stretchr/testify/assert"
)

type SlAssertions struct {
	*assert.Assertions
}

func New(t assert.TestingT) *SlAssertions {
	return &SlAssertions{assert.New(t)}
}

func (sla *SlAssertions) PortPushes(exp interface{}, p *core.Port) {
	a := p.Pull()
	sla.Equal(exp, a)
}

func (sla *SlAssertions) PortPushesAll(exp []interface{}, p *core.Port) {
	for _, e := range exp {
		a := p.Pull()
		sla.Equal(e, a)
	}
}

func (sla *SlAssertions) NoError(err error, msgAndArgs ...interface{}) bool {
	return sla.Assertions.NoError(err, msgAndArgs) || sla.FailNow(fmt.Sprintf("%s", err))
}
