package api

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Conditions slang.build.unservable-fails-by-name and
// ecosystem.api.no-silent-substitution: a built-in the profile cannot serve fails
// building by instance and capability, and a document's copy of a built-in is
// never built in its place (slang#276).

var (
	writeFileOperator = uuid.MustParse("9b61597d-cfbc-42d1-9620-210081244ba1")
	redisGetOperator  = uuid.MustParse("362482c1-2021-4e5c-9463-b580a6c1967e")
)

// sideBranchBundle passes its input to its output and, on the side, into one
// instance of operator whose output nothing reads. Like every bundle the studio
// exports, it carries a copy of that operator's blueprint.
func sideBranchBundle(t *testing.T, operator core.Blueprint) *core.SlangBundle {
	t.Helper()
	elem.Init()
	operator.Elementary = uuid.Nil // a copy read from a document does not say it is a built-in
	in := operator.ServiceDefs[core.MAIN_SERVICE].In
	id := uuid.New()
	program := core.Blueprint{
		Id: id,
		ServiceDefs: map[string]*core.ServiceDef{core.MAIN_SERVICE: {
			In: in.Copy(), Out: in.Copy(),
		}},
		InstanceDefs: core.InstanceDefList{{Name: "side", Operator: operator.Id}},
		Connections:  map[string][]string{"(": {")", "(side"}},
	}
	return &core.SlangBundle{Main: id, Blueprints: map[uuid.UUID]core.Blueprint{id: program, operator.Id: operator}}
}

func writeFileCopy(t *testing.T) core.Blueprint {
	t.Helper()
	elem.Init()
	blueprint, err := elem.GetBlueprint(writeFileOperator)
	require.NoError(t, err)
	return *blueprint
}

func TestBuildingFailsNamingAnUnservableSideBranch(t *testing.T) {
	_, err := BuildOperatorWith(sideBranchBundle(t, writeFileCopy(t)), elem.HostedProfile, elem.LocalCapabilities())
	require.EqualError(t, err, `instance side: built-in "write file" is unavailable under profile hosted/1: files.replace is not implemented`)

	_, err = BuildOperatorWith(sideBranchBundle(t, writeFileCopy(t)), elem.LocalProfile, elem.LocalCapabilities())
	require.NoError(t, err, "the same program builds where writing files is implemented")
}

// Before slang#276's fix, the embedded copy of a missing built-in was built as a
// composite with no operators, and the side branch silently did nothing.
func TestBuildingFailsNamingARemovedBuiltinDespiteItsCopy(t *testing.T) {
	redisGet := core.Blueprint{
		Id: redisGetOperator,
		ServiceDefs: map[string]*core.ServiceDef{core.MAIN_SERVICE: {
			In: core.TypeDef{Type: "string"}, Out: core.TypeDef{Type: "string"},
		}},
		Meta: core.BlueprintMetaDef{Name: "Redis Get"},
	}
	_, err := BuildOperatorWith(sideBranchBundle(t, redisGet), elem.LocalProfile, elem.LocalCapabilities())
	require.EqualError(t, err, `instance side: built-in "Redis Get" was removed after v0.1.27`)
}

// A document that defines its own blueprint under a built-in's ID gets the
// engine's built-in, never its definition.
func TestBuildingIgnoresADocumentsDefinitionOfABuiltin(t *testing.T) {
	spoof := writeFileCopy(t)
	spoof.InstanceDefs = core.InstanceDefList{{Name: "hidden", Operator: uuid.New()}}
	_, err := BuildOperatorWith(sideBranchBundle(t, spoof), elem.HostedProfile, elem.LocalCapabilities())
	require.EqualError(t, err, `instance side: built-in "write file" is unavailable under profile hosted/1: files.replace is not implemented`,
		"the spoofed composite, whose instance is unknown, must not be what gets built")
}

func TestAProgramsMainBlueprintMustBeComposite(t *testing.T) {
	bundle := sideBranchBundle(t, writeFileCopy(t))
	bundle.Main = writeFileOperator
	_, err := BuildOperatorWith(bundle, elem.LocalProfile, elem.LocalCapabilities())
	require.EqualError(t, err, "a program's main blueprint must be composite, not the built-in "+writeFileOperator.String())
}
