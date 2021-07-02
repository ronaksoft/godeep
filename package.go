package godeep

import (
	"encoding/json"
	"github.com/fatih/color"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
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

//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -dir=.

func FindPackages(allPackages *Packages, rootPath string, onDone func(path string)) error {
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
			if !info.IsDir() ||
				strings.HasPrefix(relPath, ".") ||
				strings.Contains(relPath, "vendor/",
				) {
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
				if len(pkg.Errors) > 0 || !strings.Contains(pkg.PkgPath, ".") {
					continue
				}
				allPackages.Fill(pkg)
				if onDone != nil {
					onDone(path)
				}
			}
			return
		}(path, info, err)
		return nil
	})
	waitGroup.Wait()
	for path, pkg := range allPackages.byPath {
		for iPath := range allPackages.importedBy[path] {
			pkg.importedByPackages = append(pkg.importedByPackages, iPath)
		}
	}
	return err
}

type Packages struct {
	byPath     map[string]*Package
	importedBy map[string]map[string]struct{}
	mtx        sync.RWMutex
}

func InitPackages() *Packages {
	return &Packages{
		byPath:     make(map[string]*Package),
		importedBy: make(map[string]map[string]struct{}),
		mtx:        sync.RWMutex{},
	}
}

func (a *Packages) Reset() {
	a.mtx.Lock()
	for k := range a.byPath {
		delete(a.byPath, k)
	}
	a.mtx.Unlock()
}

func (a *Packages) Unmarshal(data []byte) error {
	return json.Unmarshal(data, &a.byPath)
}

func (a *Packages) Marshal() []byte {
	d := jsonPackages{}
	for _, p := range a.byPath {
		d.Packages = append(d.Packages, jsonPackage{
			Name:       p.name,
			Path:       p.path,
			Imported:   p.imported,
			ImportedBy: p.importedByPackages,
			Funcs:      p.exportedFunctions,
			Types:      p.exportedTypes,
		})
	}
	// d, _ := json.Marshal(a.byPath)
	return []byte(d.JSON())
}

func (a *Packages) Exist(pkg *packages.Package) bool {
	a.mtx.RLock()
	p := a.byPath[pkg.PkgPath]
	a.mtx.RUnlock()
	return p != nil
}

func (a *Packages) Insert(pkg *Package) {
	a.mtx.Lock()
	a.byPath[pkg.path] = pkg
	a.mtx.Unlock()
}

func (a *Packages) GetByPath(pkgPath string) *Package {
	a.mtx.RLock()
	p := a.byPath[pkgPath]
	a.mtx.RUnlock()

	return p

}

func (a *Packages) Fill(pkg *packages.Package) {
	if a.Exist(pkg) {
		return
	}
	p := &Package{
		name: pkg.Name,
		path: pkg.PkgPath,
	}

	a.byPath[pkg.PkgPath] = p
	for _, ipkg := range pkg.Imports {
		p.imported = append(p.imported, ipkg.PkgPath)
		if a.importedBy[ipkg.PkgPath] == nil {
			a.importedBy[ipkg.PkgPath] = map[string]struct{}{}
		}
		a.importedBy[ipkg.PkgPath][pkg.PkgPath] = struct{}{}
	}

	for _, f := range pkg.Syntax {
		for _, o := range f.Scope.Objects {
			switch x := o.Decl.(type) {
			case *ast.TypeSpec:
				if x.Name.IsExported() {
					p.exportedTypes = append(p.exportedTypes, x.Name.Name)
				}
			case *ast.FuncDecl:
				if x.Name.IsExported() {
					fn := strings.Builder{}
					fn.WriteString(x.Name.Name)
					fn.WriteString(astFuncType(x.Type))
					p.exportedFunctions = append(p.exportedFunctions, fn.String())
				}
			case *ast.ValueSpec:
				for _, n := range x.Names {
					if n.IsExported() {
						p.exportedVariables = append(p.exportedVariables, n.Name)
					}
				}

			default:
			}

		}
	}
}

func (a *Packages) ForEach(f func(pkgPath string, pkg *Package)) {
	a.mtx.RLock()
	for key, pkg := range a.byPath {
		f(key, pkg)
	}
	a.mtx.RUnlock()
}

type Package struct {
	mtx                sync.Mutex
	name               string
	path               string
	imported           []string
	importedByPackages []string
	exportedTypes      []string
	exportedVariables  []string
	exportedFunctions  []string
}

func (p *Package) Print() {
	color.Green("========== %s (%s) ========", p.name, p.path)
	printPackage(p)
	printExportedItems(p)
}
func printPackage(pkg *Package) {
	color.Red("Imports: (%d)", len(pkg.imported))
	cnt := 0
	for p := range pkg.imported {
		cnt++
		color.Red("\t %d. %s", cnt, p)
	}
	color.HiBlue("imported By: (%d)", len(pkg.importedByPackages))
	cnt = 0
	for p := range pkg.importedByPackages {
		cnt++
		color.HiBlue("\t %d. %s", cnt, p)
	}
}
func printExportedItems(pkg *Package) {
	color.HiMagenta("Exported Functions: (%d)", len(pkg.exportedFunctions))
	cnt := 0
	for p := range pkg.exportedFunctions {
		cnt++
		color.HiMagenta("\t %d. %s", cnt, p)
	}
	color.HiGreen("Exported Types: (%d)", len(pkg.exportedTypes))
	cnt = 0
	for p := range pkg.exportedTypes {
		cnt++
		color.HiGreen("\t %d. %s", cnt, p)
	}
	color.HiRed("Exported Variables: (%d)", len(pkg.exportedVariables))
	cnt = 0
	for p := range pkg.exportedVariables {
		cnt++
		color.HiRed("\t %d. %s", cnt, p)
	}
}
