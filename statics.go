package main

import (
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
	"strings"
)

/*
   Creation Time: 2020 - Jan - 26
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func FindPackages(rootPath string) (map[string]*packages.Package, error) {
	allPackages := make(map[string]*packages.Package)
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		relPath, _ := filepath.Rel(rootPath, path)
		if !info.IsDir() || strings.HasPrefix(relPath, ".") {
			return nil
		}
		pkg, err := packages.Load(&packages.Config{
			Mode: packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedName,
			Dir:  path,
		}, "")
		if err != nil {
			return nil
		}
		if len(pkg) > 0 {
			allPackages[pkg[0].PkgPath] = pkg[0]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allPackages, nil
}
