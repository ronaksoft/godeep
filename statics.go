package main

import (
	"fmt"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
)

/*
   Creation Time: 2020 - Jan - 26
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/


func FindPackages(rootPath string) {
	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		pkgs, err := packages.Load(&packages.Config{}, "")
		if err != nil {
			fmt.Println(path, ":::", err)
		} else {
			fmt.Println(path, ":::", len(pkgs), "Packages")
		}
		return nil
	})
}