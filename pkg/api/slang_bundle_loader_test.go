package api

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSlangBundleLoaderListsBlueprintIDs(t *testing.T) {
	first := uuid.New()
	second := uuid.New()
	loader := &slangBundleLoader{blueprintById: map[uuid.UUID]core.Blueprint{
		first:  {Id: first},
		second: {Id: second},
	}}

	ids, err := loader.List()
	require.NoError(t, err)
	require.ElementsMatch(t, []uuid.UUID{first, second}, ids)
}
