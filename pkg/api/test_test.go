package api

import (
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
)

func TestTestEqual__Bools(t *testing.T) {
	a := assertions.New(t)
	a.True(testEqual(true, true))
	a.True(testEqual(false, false))
	a.False(testEqual(true, false))
	a.False(testEqual(false, true))
}

func TestTestEqual__Strings(t *testing.T) {
	a := assertions.New(t)
	a.False(testEqual("", nil))
	a.False(testEqual("", "test"))
	a.True(testEqual("", ""))
	a.True(testEqual("test", "test"))
}

func TestTestEqual__Number(t *testing.T) {
	a := assertions.New(t)
	a.False(testEqual(1, 0))
	a.False(testEqual(1.0, 0))
	a.False(testEqual(1, 0.0))
	a.False(testEqual(1.0, 0.0))

	a.True(testEqual(1, 1))
	a.True(testEqual(0, 0))
	a.True(testEqual(1.0, 1))
	a.True(testEqual(0.0, 0))
	a.True(testEqual(1, 1.0))
	a.True(testEqual(0, 0.0))
	a.True(testEqual(1.0, 1.0))
	a.True(testEqual(0.0, 0.0))
}

func TestTestEqual__Streams(t *testing.T) {
	a := assertions.New(t)
	a.False(testEqual([]interface{}{1}, []interface{}{0}))
	a.True(testEqual([]interface{}{1}, []interface{}{1}))
	a.True(testEqual([]interface{}{1.0}, []interface{}{1}))
	a.False(testEqual([]interface{}{1, 2}, []interface{}{1}))
	a.False(testEqual([]interface{}{1}, []interface{}{1, 2}))
}

func TestTestEqual__Maps(t *testing.T) {
	a := assertions.New(t)
	a.True(testEqual(map[string]interface{}{"a": true}, map[string]interface{}{"a": true}))
	a.False(testEqual(map[string]interface{}{"a": true, "b": true}, map[string]interface{}{"a": true}))
	a.False(testEqual(map[string]interface{}{"a": true}, map[string]interface{}{"a": true, "b": true}))

	a.False(testEqual(map[string]interface{}{"a": true}, map[string]interface{}{"a": false}))
	a.True(testEqual(map[string]interface{}{"a": 1}, map[string]interface{}{"a": 1}))
	a.True(testEqual(map[string]interface{}{"a": 1}, map[string]interface{}{"a": 1.0}))
}
