// Adapted from wgpu's tools/apigolden tests: they build packages in memory, so
// they run in milliseconds and cannot be affected by the tree under analysis.
package main

import (
	"go/constant"
	"go/token"
	"go/types"
	"testing"
)

func mustInsert(t *testing.T, pkg *types.Package, obj types.Object) {
	t.Helper()
	if alt := pkg.Scope().Insert(obj); alt != nil {
		t.Fatalf("insert %s: already declared", obj.Name())
	}
}

func signature(recv *types.Var, params, results []*types.Var) *types.Signature {
	return types.NewSignatureType(recv, nil, nil, types.NewTuple(params...), types.NewTuple(results...), false)
}

func TestFormatStructConstVarFuncAlias(t *testing.T) {
	pkg := types.NewPackage("example.com/x", "x")

	mustInsert(t, pkg, types.NewConst(token.NoPos, pkg, "MaxSize", types.Typ[types.Uint32], constant.MakeUint64(16)))
	mustInsert(t, pkg, types.NewVar(token.NoPos, pkg, "ErrX", types.Universe.Lookup("error").Type()))

	emb := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Embedded", nil), types.NewStruct(nil, nil), nil)
	mustInsert(t, pkg, emb.Obj())

	st := types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, pkg, "Name", types.Typ[types.String], false),
		types.NewField(token.NoPos, pkg, "hidden", types.Typ[types.Int], false), // unexported: must not appear
		types.NewField(token.NoPos, pkg, "Embedded", emb, true),                 // embedded: printed as the type
	}, nil)
	thing := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Thing", nil), st, nil)
	mustInsert(t, pkg, thing.Obj())

	mustInsert(t, pkg, types.NewFunc(token.NoPos, pkg, "NewThing",
		signature(nil, nil, []*types.Var{types.NewVar(token.NoPos, pkg, "", types.NewPointer(thing))})))

	// Exported method with a pointer receiver, plus an unexported one that must
	// not be recorded.
	thing.AddMethod(types.NewFunc(token.NoPos, pkg, "Size",
		signature(types.NewVar(token.NoPos, pkg, "t", types.NewPointer(thing)), nil,
			[]*types.Var{types.NewVar(token.NoPos, pkg, "", types.Typ[types.Int])})))
	thing.AddMethod(types.NewFunc(token.NoPos, pkg, "hidden",
		signature(types.NewVar(token.NoPos, pkg, "t", types.NewPointer(thing)), nil, nil)))

	// An alias must stay distinguishable from a defined type.
	aliasTN := types.NewTypeName(token.NoPos, pkg, "AliasOne", nil)
	types.NewAlias(aliasTN, thing)
	mustInsert(t, pkg, aliasTN)

	got := FormatPackage(pkg)
	want := `package example.com/x

type AliasOne = Thing   # alias

type Embedded struct {
}

var ErrX error

const MaxSize uint32 = 16

func NewThing() *Thing

type Thing struct {
    Name string
    Embedded
}
method (*Thing) Size() int
`
	if got != want {
		t.Fatalf("FormatPackage mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFormatInterfaceKeywordAndEmbedded(t *testing.T) {
	pkg := types.NewPackage("example.com/y", "y")

	iface := types.NewInterfaceType([]*types.Func{
		types.NewFunc(token.NoPos, pkg, "Zeta", signature(nil, nil, nil)),
		types.NewFunc(token.NoPos, pkg, "Alpha", signature(nil, nil,
			[]*types.Var{types.NewVar(token.NoPos, pkg, "", types.Universe.Lookup("error").Type())})),
	}, nil)
	iface.Complete()
	mustInsert(t, pkg, types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Doer", nil), iface, nil).Obj())

	got := FormatPackage(pkg)
	want := `package example.com/y

type Doer interface {
    Alpha() error
    Zeta()
}
`
	if got != want {
		t.Fatalf("interface mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// The golden is only useful if two runs agree: the same declarations inserted in
// a different order must produce the same bytes.
func TestSortingIsOrderIndependent(t *testing.T) {
	build := func(names []string) string {
		pkg := types.NewPackage("example.com/z", "z")
		for _, n := range names {
			mustInsert(t, pkg, types.NewVar(token.NoPos, pkg, n, types.Typ[types.Int]))
		}
		return FormatPackage(pkg)
	}
	forward := build([]string{"Alpha", "Beta", "Gamma"})
	reverse := build([]string{"Gamma", "Beta", "Alpha"})
	if forward != reverse {
		t.Fatalf("output depends on insertion order:\n--- forward ---\n%s\n--- reverse ---\n%s", forward, reverse)
	}
}

// Unexported identifiers and unexported methods never reach the golden.
func TestUnexportedIsExcluded(t *testing.T) {
	pkg := types.NewPackage("example.com/w", "w")
	mustInsert(t, pkg, types.NewVar(token.NoPos, pkg, "hidden", types.Typ[types.Int]))
	got := FormatPackage(pkg)
	want := "package example.com/w\n"
	if got != want {
		t.Fatalf("unexported leaked:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
