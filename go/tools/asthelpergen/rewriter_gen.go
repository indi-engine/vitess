/*
Copyright 2021 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"go/types"

	"github.com/dave/jennifer/jen"
)

type rewriterGen struct {
	cases           []jen.Code
	replaceMethods  []jen.Code
	interestingType func(types.Type) bool
	ifaceName       string
}

func newRewriterGen(f func(types.Type) bool, name string) rewriterGen {
	return rewriterGen{interestingType: f, ifaceName: name}
}

var noQualifier = func(p *types.Package) string {
	return ""
}

func (r *rewriterGen) visitStruct(typeString, replaceMethodPrefix string, stroct *types.Struct, pointer bool) error {
	var caseStmts []jen.Code
	for i := 0; i < stroct.NumFields(); i++ {
		field := stroct.Field(i)
		if r.interestingType(field.Type()) {
			if pointer {
				replacerName, method := r.createReplaceMethod(replaceMethodPrefix, typeString, field)
				r.replaceMethods = append(r.replaceMethods, method)

				caseStmts = append(caseStmts, caseStmtFor(field, replacerName))
			} else {
				caseStmts = append(caseStmts, casePanicStmtFor(field, typeString+" "+field.Name()))
			}
		}
		sliceT, ok := field.Type().(*types.Slice)
		if ok && r.interestingType(sliceT.Elem()) { // we have a field containing a slice of interesting elements
			replaceMethodName := replaceMethodPrefix
			if !pointer {
				replaceMethodName += "Val"
			}
			replacerName, methods := r.createReplaceCodeForSliceField(replaceMethodName, typeString, field)
			r.replaceMethods = append(r.replaceMethods, methods...)
			caseStmts = append(caseStmts, caseStmtForSliceField(field, replacerName)...)
		}
	}
	r.cases = append(r.cases, jen.Case(jen.Id(typeString)).Block(caseStmts...))
	return nil
}

func (r *rewriterGen) visitSlice(t types.Type, typeString, replaceMethodPrefix string, slice *types.Slice) error {
	name, replaceMethod := r.createReplaceCodeForSlice(replaceMethodPrefix, typeString, types.TypeString(slice.Elem(), noQualifier))
	r.replaceMethods = append(r.replaceMethods, replaceMethod)
	r.cases = append(r.cases, jen.Case(jen.Id(typeString)).Block(caseStmtForSlice(name)))
	return nil
}

func caseStmtFor(field *types.Var, name string) *jen.Statement {
	return jen.Id("a").Dot("apply").Call(jen.Id("node"), jen.Id("n").Dot(field.Name()), jen.Id(name))
}

func casePanicStmtFor(field *types.Var, name string) *jen.Statement {
	return jen.Id("a").Dot("apply").Call(jen.Id("node"), jen.Id("n").Dot(field.Name()), jen.Id("replacePanic").Call(jen.Lit(name)))
}

func caseStmtForSlice(name string) jen.Code {
	return jen.For(jen.List(jen.Op("x"), jen.Id("el"))).Op(":=").Range().Id("n").Block(
		jen.Id("a").Dot("apply").Call(
			jen.Id("node"),
			jen.Id("el"),
			jen.Id(name).Call(jen.Id("x")),
		),
	)
}

func caseStmtForSliceField(field *types.Var, name string) []jen.Code {
	return []jen.Code{
		jen.For(jen.List(jen.Op("x"), jen.Id("el"))).Op(":=").Range().Id("n").Dot(field.Name()).Block(
			jen.Id("a").Dot("apply").Call(
				jen.Id("node"),
				jen.Id("el"),
				jen.Id(name).Call(jen.Id("x")),
			),
		),
	}
}

func (r *rewriterGen) structCase(name string, stroct *types.Struct) (jen.Code, error) {
	var stmts []jen.Code
	for i := 0; i < stroct.NumFields(); i++ {
		field := stroct.Field(i)
		if r.interestingType(field.Type()) {
			stmts = append(stmts, jen.Id("a").Dot("apply").Call(jen.Id("node"), jen.Id("n").Dot(field.Name()), jen.Nil()))
		}
	}
	return jen.Case(jen.Op("*").Id(name)).Block(stmts...), nil
}

func (r *rewriterGen) createReplaceMethod(structName, structType string, field *types.Var) (string, jen.Code) {
	name := "replace" + structName + field.Name()
	return name, jen.Func().Id(name).Params(
		jen.Id("newNode"),
		jen.Id("parent").Id(r.ifaceName),
	).Block(
		jen.Id("parent").Assert(jen.Id(structType)).Dot(field.Name()).Op("=").Id("newNode").Assert(jen.Id(types.TypeString(field.Type(), noQualifier))),
	)
}

func (r *rewriterGen) createReplaceCodeForSlice(structName, structType, elemType string) (string, jen.Code) {
	name := "replace" + structName

	/*
		func replacer(idx int) func(AST, AST) {
			return func(newnode, container AST) {
				container.(InterfaceSlice)[idx] = newnode.(AST)
			}
		}

	*/

	s := jen.Func().Id(name).Params(jen.Id("idx").Int()).Func().Params(jen.List(jen.Id("AST"), jen.Id("AST"))).Block(
		jen.Return(jen.Func().Params(jen.List(jen.Id("newNode"), jen.Id("container")).Id("AST"))).Block(
			jen.Id("container").Assert(jen.Id(structType)).Index(jen.Id("idx")).Op("=").
				Id("newNode").Assert(jen.Id(elemType)),
		),
	)

	return name, s
}

func (r *rewriterGen) createReplaceCodeForSliceField(structName, structType string, field *types.Var) (string, []jen.Code) {
	name := "replace" + structName + field.Name()
	elemType := field.Type().(*types.Slice).Elem()

	/*
		func replacerStructField(idx int) func(AST, AST) {
			return func(newNode, container AST) {
				container.(*Struct)[idx] = newNode.(AST)
			}
		}

	*/

	s := jen.Func().Id(name).Params(jen.Id("idx").Int()).Func().Params(jen.List(jen.Id("AST"), jen.Id("AST"))).Block(
		jen.Return(jen.Func().Params(jen.List(jen.Id("newNode"), jen.Id("container")).Id("AST"))).Block(
			jen.Id("container").Assert(jen.Id(structType)).Dot(field.Name()).Index(jen.Id("idx")).Op("=").
				Id("newNode").Assert(jen.Id(types.TypeString(elemType, noQualifier))),
		),
	)

	return name, []jen.Code{
		s,
	}
}

func (r *rewriterGen) createFile(pkgName string) *jen.File {
	out := jen.NewFile(pkgName)
	out.HeaderComment(licenseFileHeader)
	out.HeaderComment("Code generated by ASTHelperGen. DO NOT EDIT.")

	for _, method := range r.replaceMethods {
		out.Add(method)
	}

	out.Add(r.applyFunc())

	return out
}

func (r *rewriterGen) applyFunc() *jen.Statement {
	// func (a *application) apply(parent, node SQLNode, replacer replacerFunc) {
	apply := jen.Func().Params(
		jen.Id("a").Op("*").Id("application"),
	).Id("apply").Params(
		jen.Id("parent"),
		jen.Id("node").Id(r.ifaceName),
		jen.Id("replacer").Id("replacerFunc"),
	).Block(
		/*
			if node == nil || isNilValue(node) {
				return
			}
		*/
		jen.If(
			jen.Id("node").Op("==").Nil().Op("||").
				Id("isNilValue").Call(jen.Id("node"))).Block(
			jen.Return(),
		),
		/*
			saved := a.cursor
			a.cursor.replacer = replacer
			a.cursor.node = node
			a.cursor.parent = parent
		*/
		jen.Id("saved").Op(":=").Id("a").Dot("cursor"),
		jen.Id("a").Dot("cursor").Dot("replacer").Op("=").Id("replacer"),
		jen.Id("a").Dot("cursor").Dot("node").Op("=").Id("node"),
		jen.Id("a").Dot("cursor").Dot("parent").Op("=").Id("parent"),
		jen.If(
			jen.Id("a").Dot("pre").Op("!=").Nil().Op("&&").
				Op("!").Id("a").Dot("pre").Call(jen.Op("&").Id("a").Dot("cursor"))).Block(
			jen.Id("a").Dot("cursor").Op("=").Id("saved"),
			jen.Return(),
		),

		//	switch n := node.(type) {
		jen.Switch(jen.Id("n").Op(":=").Id("node").Assert(jen.Id("type")).Block(
			r.cases...,
		)),

		/*
			if a.post != nil && !a.post(&a.cursor) {
				panic(abort)
			}
		*/
		jen.If(
			jen.Id("a").Dot("post").Op("!=").Nil().Op("&&").
				Op("!").Id("a").Dot("post").Call(jen.Op("&").Id("a").Dot("cursor"))).Block(
			jen.Id("panic").Call(jen.Id("abort")),
		),

		// 	a.cursor = saved
		jen.Id("a").Dot("cursor").Op("=").Id("saved"),
	)
	return apply
}
