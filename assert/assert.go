package assert

import (
	"reflect"
	"slang/op"
	"testing"
)

func PortItems(t *testing.T, i []interface{}, p *op.Port) {
	t.Helper()
	for _, e := range i {
		a := p.Pull()
		if !reflect.DeepEqual(e, a) {
			t.Errorf("wrong value:\nexpected: %#v,\nactual:   %#v", e, a)
			break
		}
	}
}
