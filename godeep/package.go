package godeep

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

/*
   Creation Time: 2020 - Jan - 26
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func FindPackages(allPackages *Packages, rootPath string) error {
	allPackages.Reset()
	waitGroup := sync.WaitGroup{}
	rateLimit := make(chan struct{}, 50)
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		waitGroup.Add(1)
		rateLimit <- struct{}{}
		go func(path string, info os.FileInfo, err error) {
			defer waitGroup.Done()
			defer func() {
				<-rateLimit
			}()
			relPath, _ := filepath.Rel(rootPath, path)
			if !info.IsDir() || strings.HasPrefix(relPath, ".") {
				return
			}
			pkgs, err := packages.Load(&packages.Config{
				Mode: packages.NeedImports | packages.NeedDeps | packages.NeedTypes |
					packages.NeedName | packages.NeedSyntax,
				Dir: path,
			}, "")
			if err != nil {
				return
			}
			for _, pkg := range pkgs {
				allPackages.Fill(pkg)
				fmt.Println(fmt.Sprintf("Package '%s' analyzed.", path))
			}
			return
		}(path, info, err)
		return nil
	})
	waitGroup.Wait()
	return err
}

type Packages struct {
	m   map[string]*Package
	mtx sync.RWMutex
}

func InitPackages() *Packages {
	return &Packages{
		m:   make(map[string]*Package),
		mtx: sync.RWMutex{},
	}
}

func (a *Packages) Reset() {
	a.mtx.Lock()
	for k := range a.m {
		delete(a.m, k)
	}
	a.mtx.Unlock()
}

func (a *Packages) Unmarshal(data []byte) error {
	return json.Unmarshal(data, &a.m)
}

func (a *Packages) Marshal() []byte {
	d, _ := json.Marshal(a.m)
	return d
}

func (a *Packages) GetByPackage(pkg *packages.Package) *Package {
	p := a.GetByPath(pkg.PkgPath)
	if p == nil {
		p = &Package{
			Name: pkg.Name,
			Path: pkg.PkgPath,
		}
		a.mtx.Lock()
		a.m[pkg.PkgPath] = p
		a.mtx.Unlock()
	}
	return p
}

func (a *Packages) GetByPath(pkgPath string) *Package {
	a.mtx.RLock()
	p := a.m[pkgPath]
	a.mtx.RUnlock()

	return p

}

func (a *Packages) Fill(pkg *packages.Package) {
	p := a.GetByPackage(pkg)
	if len(p.DirectImportedPackages) == 0 {
		for _, ipkg := range pkg.Imports {
			p.Import(a.GetByPackage(ipkg))
			a.Fill(ipkg)
		}
	}
	for _, f := range pkg.Syntax {
		for n, o := range f.Scope.Objects {
			switch x := o.Decl.(type) {
			case *ast.TypeSpec:
				if x.Name.IsExported() {
					p.ExportedType(x.Name.Name)
				}
			case *ast.FuncDecl:
				if x.Name.IsExported() {
					fn := strings.Builder{}
					fn.WriteString(x.Name.Name)
					fn.WriteRune('(')
					for idx, p := range x.Type.Params.List {
						for idx, n := range p.Names {
							fn.WriteString(n.Name)
							if idx < len(p.Names)-1 {
								fn.WriteString(", ")
							}
						}
						fn.WriteRune(' ')
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
							fn.WriteString(", ")
						}
					}
					fn.WriteRune(')')
					p.ExportedFunc(fn.String())
				}

			case *ast.ValueSpec:
				for _, n := range x.Names {
					if n.IsExported() {
						p.ExportedVar(n.Name)
					}
				}

			default:
				fmt.Println(n, reflect.TypeOf(o.Decl))
			}

		}
	}
}

func (a *Packages) ForEach(f func(pkgPath string, pkg *Package)) {
	a.mtx.RLock()
	for key, pkg := range a.m {
		if pkg != nil {
			f(key, pkg)
		}
	}
	a.mtx.RUnlock()
}

type Package struct {
	mtx                    sync.Mutex
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

func (p *Package) Import(pkg *Package) {
	p.mtx.Lock()
	p.DirectImportedPackages = append(p.DirectImportedPackages, pkg.Path)
	p.DirectImports++
	p.mtx.Unlock()

	pkg.mtx.Lock()
	pkg.ImportedBy++
	pkg.ImportedByPackages = append(pkg.ImportedByPackages, pkg.Path)
	pkg.mtx.Unlock()
}

func (p *Package) ExportedFunc(name string) {
	p.mtx.Lock()
	p.ExportedFunctions = append(p.ExportedFunctions, name)
	p.mtx.Unlock()
}

func (p *Package) ExportedType(name string) {
	p.mtx.Lock()
	p.ExportedTypes = append(p.ExportedTypes, name)
	p.mtx.Unlock()
}

func (p *Package) ExportedVar(name string) {
	p.mtx.Lock()
	p.ExportedVariables = append(p.ExportedVariables, name)
	p.mtx.Unlock()
}

func (p *Package) Print() {
	color.Green("==== %s ====", p.Name)
	printPackage(p)
	printExportedItems(p)
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
