// Package buildinfo provides utilities for printing build information.
package buildinfo

import "fmt"

// Info contains build metadata set at compile time via ldflags.
type Info struct {
	Version string
	Date    string
	Commit  string
}

// valueOrNA returns "N/A" if the input string is empty, otherwise returns the original string.
func valueOrNA(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}

// Print outputs the build information to stdout.
func (i Info) Print() {
	fmt.Printf("Build version: %s\n", valueOrNA(i.Version))
	fmt.Printf("Build date: %s\n", valueOrNA(i.Date))
	fmt.Printf("Build commit: %s\n", valueOrNA(i.Commit))
}
