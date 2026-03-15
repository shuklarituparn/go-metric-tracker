// Package testdata is used to generate the reset structs
package testdata

// generate:reset
type TestStruct struct {
	ID          int
	Name        string
	Count       int64
	Active      bool
	Tags        []string
	Metadata    map[string]string
	Score       float64
	Description *string
}

// generate:reset
type NestedStruct struct {
	Parent *TestStruct
	Items  []int
	Config map[string]int
}

// generate:reset
type ComplexStruct struct {
	IDs       []int
	Names     []string
	Data      map[string]interface{}
	Counter   int
	IsEnabled bool
	Child     *NestedStruct
}

// This struct should NOT get a Reset method (no comment)
type IgnoredStruct struct {
	Value int
}
