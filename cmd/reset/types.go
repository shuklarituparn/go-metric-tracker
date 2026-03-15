package main

import "go/ast"

type StructInfo struct {
	Name   string
	Fields []*ast.Field
}

type PackageInfo struct {
	Name string
	Path string

	Structs []StructInfo

	Imports map[string]string

	NeededAliases map[string]bool
}
