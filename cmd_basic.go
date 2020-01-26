package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	RootCmd.AddCommand(CmdAnalyze, CmdPrint, CmdImport, CmdExport, CmdFunctions)
}

func ResetCommands() {
	CmdPrint.ResetCommands()
	for key := range AllPackages {
		CmdPrint.AddCommand(&cobra.Command{
			Use: key,
			Run: func(cmd *cobra.Command, args []string) {
				pkg := AllPackages[cmd.Use]
				if pkg == nil {
					return
				}
				printPackage(pkg)
				printExportedItems(pkg)
			},
		})
	}
}

var CmdImport = &cobra.Command{
	Use:   "import",
	Short: "import is used to use already exported data from analyze",
	Run: func(cmd *cobra.Command, args []string) {
		inputDir, err := cmd.Flags().GetString(FlagInputDir)
		PrintOnErr(err)

		f, err := os.Open(filepath.Join(inputDir, "all_packages.json"))
		PanicOnErr(err)
		data, err := ioutil.ReadAll(f)
		PanicOnErr(err)
		err = json.Unmarshal(data, &AllPackages)
		PanicOnErr(err)
		err = f.Close()
		PanicOnErr(err)
		ResetCommands()
	},
}

var CmdExport = &cobra.Command{
	Use:   "export",
	Short: "exports the analyzed data as a json file",
	Run: func(cmd *cobra.Command, args []string) {
		outputDir, err := cmd.Flags().GetString(FlagOutputDir)
		PrintOnErr(err)

		exportData, err := json.Marshal(AllPackages)
		PanicOnErr(err)
		f, err := os.Create(filepath.Join(outputDir, "all_packages.json"))
		PanicOnErr(err)
		_, err = f.Write(exportData)
		PanicOnErr(err)
		err = f.Close()
		PanicOnErr(err)
	},
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
			fillAllExportedItems(path, pkg.PkgPath)
			fmt.Println(fmt.Sprintf("Package '%s' analyzed.", path))
		}
		ResetCommands()

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
func fillAllExportedItems(dirPath, pkgPath string) {
	fmt.Println(dirPath, pkgPath)
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, dirPath, nil, parser.AllErrors)
	if err != nil {
		PrintOnErr(err)
	}

	pkg := AllPackages[pkgPath]
	if pkg == nil {
		color.Red("Package not found: %s", pkgPath)
		return
	}
	for _, astPkg := range pkgs {
		for _, f := range astPkg.Files {
			for n, o := range f.Scope.Objects {
				switch x := o.Decl.(type) {
				case *ast.TypeSpec:
					if x.Name.IsExported() {
						pkg.ExportedTypes = append(pkg.ExportedTypes, x.Name.Name)
					}
				case *ast.FuncDecl:
					if x.Name.IsExported() {
						fn := strings.Builder{}
						fn.WriteString(x.Name.Name)
						fn.WriteRune('(')
						for idx, p := range x.Type.Params.List {
							for idx, n := range p.Names {
								fn.WriteString(n.Name)
								fn.WriteRune(' ')
								if idx < len(p.Names)-1 {
									fn.WriteRune(',')
								}
							}
							switch xx := p.Type.(type) {
							case *ast.Ident:
								fn.WriteString(xx.Name)
							case *ast.Ellipsis:
								fn.WriteString("...")
								switch xxx := xx.Elt.(type) {
								case *ast.InterfaceType:
									fn.WriteString("interface{}")
								case *ast.Ident:
									fn.WriteString(xxx.Name)
								}
							}
							if idx < len(x.Type.Params.List)-1 {
								fn.WriteRune(',')
							}
						}
						fn.WriteRune(')')
						pkg.ExportedFunctions = append(pkg.ExportedFunctions, fn.String())
					}

				case *ast.ValueSpec:
					for _, n := range x.Names {
						if n.IsExported() {
							pkg.ExportedVariables = append(pkg.ExportedVariables, n.Name)
						}
					}

				default:
					fmt.Println(n, reflect.TypeOf(o.Decl))
				}

			}
		}
	}
}
func printPackage(pkg *Package) {
	color.Green("%s:", pkg.Path)
	color.Red("Imports: (%d)", pkg.DirectImports)
	cnt := 0
	for _, p := range pkg.DirectImportedPackages {
		cnt++
		color.Red("\t %d. %s", cnt, p)
	}
	color.HiBlue("Imported By: (%d)", pkg.ImportedBy)
	cnt = 0
	for _, p := range pkg.ImportedByPackages {
		cnt++
		color.HiBlue("\t %d. %s", cnt, p)
	}
}
func printExportedItems(pkg *Package) {
	color.HiMagenta("Exported Functions: (%d)", len(pkg.ExportedFunctions))
	cnt := 0
	for _, p := range pkg.ExportedFunctions {
		cnt++
		color.HiMagenta("\t %d. %s", cnt, p)
	}
	color.HiGreen("Exported Types: (%d)", len(pkg.ExportedTypes))
	cnt = 0
	for _, p := range pkg.ExportedTypes {
		cnt++
		color.HiGreen("\t %d. %s", cnt, p)
	}
	color.HiRed("Exported Variables: (%d)", len(pkg.ExportedVariables))
	cnt = 0
	for _, p := range pkg.ExportedVariables {
		cnt++
		color.HiRed("\t %d. %s", cnt, p)
	}
}

var CmdFunctions = &cobra.Command{
	Use: "functions",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()
		fillAllExportedItems(cwd, "")
	},
}

var CmdPrint = &cobra.Command{
	Use: "print",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			pkg, ok := AllPackages[args[0]]
			if ok {
				printPackage(pkg)
				printExportedItems(pkg)
			} else {
				fmt.Println("Package not found:", args[0])
			}
		} else {
			for _, pkg := range AllPackages {
				color.Green("==== %s ====", pkg.Name)
				printPackage(pkg)
				printExportedItems(pkg)
			}
		}

	},
}
