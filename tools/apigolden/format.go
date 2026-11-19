// Adapted from wgpu's tools/apigolden (the wgpu v0.35.0 window tooling); kept
// format-identical so both repos' goldens are reviewed the same way.
package main

import (
	"fmt"
	"go/types"
	"sort"
	"strings"
)

// FormatPackage renders the exported surface of one package as canonical text.
// It is pure: it reads a *types.Package and writes a string, so the formatter can
// be tested without loading any module.
//
// Layout per package:
//
//	package <path>
//
//	const Name Type = value
//	var  Name Type
//	type Name <underlying>            (a struct or interface block when applicable)
//	method (Recv) Name(sig)
//	func Name(sig)
//
// Identifiers are sorted by name; struct fields keep declaration order; interface
// methods are sorted by name. Doc comments are never emitted.
func FormatPackage(pkg *types.Package) string {
	q := qualifier(pkg)

	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n", pkg.Path())

	names := append([]string(nil), pkg.Scope().Names()...)
	sort.Strings(names)

	for _, name := range names {
		obj := pkg.Scope().Lookup(name)
		if obj == nil || !obj.Exported() {
			continue
		}
		switch o := obj.(type) {
		case *types.Const:
			fmt.Fprintf(&b, "\nconst %s %s = %s\n", o.Name(), TypeString(o.Type(), q), o.Val().ExactString())
		case *types.Var:
			fmt.Fprintf(&b, "\nvar %s %s\n", o.Name(), TypeString(o.Type(), q))
		case *types.Func:
			fmt.Fprintf(&b, "\nfunc %s%s\n", o.Name(), signatureBody(o.Type(), q))
		case *types.TypeName:
			b.WriteString("\n")
			b.WriteString(formatTypeName(o, q))
		}
	}
	return b.String()
}

// qualifier prints a type from another package by its full import path, so the
// golden never depends on an import alias choice; the package under analysis is
// printed unqualified.
func qualifier(self *types.Package) types.Qualifier {
	return func(other *types.Package) string {
		if other == nil || other == self {
			return ""
		}
		return other.Path()
	}
}

// TypeString is types.TypeString with the canonical qualifier applied.
func TypeString(t types.Type, q types.Qualifier) string { return types.TypeString(t, q) }

// signatureBody renders "(params) (results)" without the leading "func".
func signatureBody(t types.Type, q types.Qualifier) string {
	sig, ok := t.(*types.Signature)
	if !ok {
		return " " + TypeString(t, q)
	}
	s := types.TypeString(sig, q) // "func(a T, b U) (R, error)"
	return strings.TrimPrefix(s, "func")
}

func formatTypeName(o *types.TypeName, q types.Qualifier) string {
	var b strings.Builder

	if o.IsAlias() {
		// An alias is a distinct declaration form and must stay distinguishable
		// from a defined type: `type X = Y` does not create a new type.
		fmt.Fprintf(&b, "type %s = %s   # alias\n", o.Name(), TypeString(types.Unalias(o.Type()), q))
		return b.String()
	}

	named, ok := o.Type().(*types.Named)
	if !ok {
		fmt.Fprintf(&b, "type %s %s\n", o.Name(), TypeString(o.Type(), q))
		return b.String()
	}

	head := "type " + o.Name() + typeParams(named, q)

	switch u := named.Underlying().(type) {
	case *types.Struct:
		fmt.Fprintf(&b, "%s struct {\n", head)
		for i := 0; i < u.NumFields(); i++ {
			f := u.Field(i)
			if !f.Exported() {
				continue
			}
			if f.Embedded() {
				fmt.Fprintf(&b, "    %s\n", TypeString(f.Type(), q))
				continue
			}
			fmt.Fprintf(&b, "    %s %s\n", f.Name(), TypeString(f.Type(), q))
		}
		b.WriteString("}\n")
	case *types.Interface:
		fmt.Fprintf(&b, "%s interface {\n", head)
		for i := 0; i < u.NumExplicitMethods(); i++ {
			m := u.ExplicitMethod(i)
			if !m.Exported() {
				continue
			}
			fmt.Fprintf(&b, "    %s%s\n", m.Name(), signatureBody(m.Type(), q))
		}
		for i := 0; i < u.NumEmbeddeds(); i++ {
			fmt.Fprintf(&b, "    %s\n", TypeString(u.EmbeddedType(i), q))
		}
		b.WriteString("}\n")
	default:
		fmt.Fprintf(&b, "%s %s\n", head, TypeString(u, q))
	}

	// Explicit methods declared on this named type, exported only, sorted by name.
	type method struct {
		name string
		line string
	}
	var methods []method
	for i := 0; i < named.NumMethods(); i++ {
		m := named.Method(i)
		if !m.Exported() {
			continue
		}
		recv := o.Name()
		if sig, ok := m.Type().(*types.Signature); ok && sig.Recv() != nil {
			if _, isPtr := sig.Recv().Type().(*types.Pointer); isPtr {
				recv = "*" + recv
			}
		}
		methods = append(methods, method{
			name: m.Name(),
			line: fmt.Sprintf("method (%s) %s%s\n", recv, m.Name(), signatureBody(m.Type(), q)),
		})
	}
	sort.Slice(methods, func(i, j int) bool { return methods[i].name < methods[j].name })
	for _, m := range methods {
		b.WriteString(m.line)
	}

	return b.String()
}

func typeParams(named *types.Named, q types.Qualifier) string {
	tp := named.TypeParams()
	if tp == nil || tp.Len() == 0 {
		return ""
	}
	parts := make([]string, 0, tp.Len())
	for i := 0; i < tp.Len(); i++ {
		p := tp.At(i)
		parts = append(parts, p.Obj().Name()+" "+TypeString(p.Constraint(), q))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
