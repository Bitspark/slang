package storage

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// A built-in's ID is reserved: no document defines it, whether it is saved,
// listed or loaded (API design 5.6).

var valueOperator = uuid.MustParse("8b62495a-e482-4a3e-8020-0ab8a350ad2d")

func TestSavingUnderABuiltinsIDIsRefused(t *testing.T) {
	elem.Init()
	st := NewStorage().AddBackend(NewWritableFileSystem(t.TempDir()))
	_, err := st.Save(core.Blueprint{Id: valueOperator})
	require.EqualError(t, err, valueOperator.String()+" is the ID of a built-in and cannot be saved")
}

func TestACopyOfABuiltinIsNeitherListedNorLoaded(t *testing.T) {
	elem.Init()
	backend := NewWritableFileSystem(t.TempDir())
	copyOfValue, err := elem.GetBlueprint(valueOperator)
	require.NoError(t, err)
	copyOfValue.Meta.Name = "a copy with another name"
	_, err = backend.Save(*copyOfValue) // as an older daemon did
	require.NoError(t, err)

	st := NewStorage().AddBackend(backend)
	listed, err := st.List()
	require.NoError(t, err)
	require.NotContains(t, listed, valueOperator)

	loaded, err := st.Load(valueOperator)
	require.NoError(t, err)
	require.Equal(t, "value", loaded.Meta.Name, "the engine's built-in, not the saved copy")
}
