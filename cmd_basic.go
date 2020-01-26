package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
	"os"
)

/*
   Creation Time: 2019 - Oct - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func init() {

}

func ResetSubCommands() {
	RootCmd.AddCommand(CmdAnalyze, CmdPrint)
}

var CmdAnalyze = &cobra.Command{
	Use: "analyze",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Please be patient, this may take a bit longer than you think ...")
		cwd, _ := os.Getwd()
		pkgs, err := FindPackages(cwd)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("All Packages have been traversed, now we are building the relation")
		AllPackages = make(map[string]*Package)
		for path, pkg := range pkgs {
			fillAllPackages(pkg, nil)
			fmt.Println(fmt.Sprintf("Package '%s' analyzed.", path))
		}
		CmdPrint.ResetCommands()
		for key := range AllPackages {
			CmdPrint.AddCommand(&cobra.Command{
				Use: key,
				Run: func(cmd *cobra.Command, args []string) {
					pkg := AllPackages[cmd.Use]
					color.Green("%s:", cmd.Use)
					color.Red("Imported By: %d,  Direct Imports: %d", pkg.ImportedBy, pkg.DirectImports)
					cnt := 0
					for _, p := range pkg.ImportedByPackages {
						cnt++
						color.HiBlue("\t\t %d. %s", cnt, p)
					}
				},
			})
		}
	},
}
func getPackage(pkg *packages.Package) *Package {
	p, ok := AllPackages[pkg.PkgPath]
	if !ok {
		p = &Package{
			Name: pkg.Name,
			Path: pkg.PkgPath,
		}
		AllPackages[pkg.PkgPath] = p
	}
	return p
}
func fillAllPackages(pkg *packages.Package, parents []string) {
	p := getPackage(pkg)
	if len(p.DirectImportedPackages) == 0 {
		p.DirectImports = len(pkg.Imports)
		for _, ipkg := range pkg.Imports {
			p.DirectImportedPackages = append(p.DirectImportedPackages, ipkg.PkgPath)
			imported := getPackage(ipkg)
			imported.ImportedBy++
			imported.ImportedByPackages = append(imported.ImportedByPackages, pkg.PkgPath)
			fillAllPackages(ipkg, append(parents, pkg.PkgPath))
		}
	}
}

var CmdPrint = &cobra.Command{
	Use: "print",
	Run: func(cmd *cobra.Command, args []string) {
		for path, pkg := range AllPackages {
			color.Green("%s:", path)
			color.Red("Imported By: %d,  Direct Imports: %d", pkg.ImportedBy, pkg.DirectImports)
			cnt := 0
			for _, p := range pkg.ImportedByPackages {
				cnt++
				color.HiBlue("\t\t %d. %s", cnt, p)
			}
		}
	},
}
