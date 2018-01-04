package builtin

import "errors"

func EnsureThat(cond bool, errMsg string) error {
	if !cond {
		return errors.New(errMsg)
	}
	return nil
}
