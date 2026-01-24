package main

import (
	"log"
	"os"
)

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	packages, err := scanPackages(rootDir)
	if err != nil {
		log.Fatalf("Failed to scan packages: %v", err)
	}

	for _, pkg := range packages {
		if len(pkg.Structs) == 0 {
			continue
		}
		if err := generateResetFile(pkg); err != nil {
			log.Printf("Failed to generate reset file for package %s: %v", pkg.Name, err)
		} else {
			log.Printf("Generated reset.gen.go for package %s with %d struct(s)", pkg.Name, len(pkg.Structs))
		}
	}
}
