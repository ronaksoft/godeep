package main

/*
   Creation Time: 2020 - Jan - 26
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

var AllPackages map[string]*Package

type Package struct {
	Name                   string
	Path                   string
	DirectImports          int
	TotalImports           int
	ImportedBy             int
	DirectImportedPackages []string
	ImportedPackages       []string
	ImportedByPackages     []string
	ExportedTypes          []string
	ExportedVariables      []string
	ExportedFunctions      []string
}
