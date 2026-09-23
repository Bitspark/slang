package main

import (
	"testing"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRunRejectsUnconnectedOutputsBeforeStarting(t *testing.T) {
	for _, mode := range SupportedRunModes {
		t.Run(mode, func(t *testing.T) {
			elem.Init()
			id := uuid.New()
			blueprint := core.Blueprint{
				Id: id,
				ServiceDefs: map[string]*core.ServiceDef{core.MAIN_SERVICE: {
					In: core.TypeDef{Type: "number"},
					Out: core.TypeDef{Type: "map", Map: core.TypeDefMap{
						"a": {Type: "number"}, "b": {Type: "number"},
					}},
				}},
			}
			o, err := api.BuildOperator(&core.SlangBundle{
				Main: id, Blueprints: map[uuid.UUID]core.Blueprint{id: blueprint},
			})
			require.NoError(t, err)
			require.EqualError(t, run(o, mode, "invalid-bind-address"),
				"unconnected output ports: )a, )b")
		})
	}
}

func TestRunRejectsMissingMainService(t *testing.T) {
	o, err := core.NewOperator("empty", nil, nil, nil, nil, core.Blueprint{})
	require.NoError(t, err)
	require.EqualError(t, run(o, "process", ""), "blueprint has no main service")
}

func TestSafeModeFromEnv(t *testing.T) {
	for value, want := range map[string]bool{"": false, "false": false, "0": false, "true": true, "1": true} {
		got, err := safeModeFromEnv(value)
		require.NoError(t, err, value)
		require.Equal(t, want, got, value)
	}
	_, err := safeModeFromEnv("yes")
	require.Error(t, err, "an unreadable value must not fall back to unsafe mode")
}
