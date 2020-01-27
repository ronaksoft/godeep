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
				if onDone != nil {
					onDone(path)
				}
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
			Name:                   pkg.Name,
			Path:                   pkg.PkgPath,
			DirectImportedPackages: make(map[string]struct{}),
			ImportedByPackages:     make(map[string]struct{}),
			ImportedPackages:       make(map[string]struct{}),
			ExportedVariables:      make(map[string]struct{}),
			ExportedFunctions:      make(map[string]struct{}),
			ExportedTypes:          make(map[string]struct{}),
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
		for _, o := range f.Scope.Objects {
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
	DirectImportedPackages map[string]struct{}
	ImportedPackages       map[string]struct{}
	ImportedByPackages     map[string]struct{}
	ExportedTypes          map[string]struct{}
	ExportedVariables      map[string]struct{}
	ExportedFunctions      map[string]struct{}
}

func (p *Package) Import(pkg *Package) {
	p.mtx.Lock()
	p.DirectImportedPackages[pkg.Path] = struct{}{}
	p.mtx.Unlock()

	pkg.mtx.Lock()
	pkg.ImportedByPackages[p.Path] = struct{}{}
	pkg.mtx.Unlock()
}

func (p *Package) ExportedFunc(name string) {
	p.mtx.Lock()
	p.ExportedFunctions[name] = struct{}{}
	p.mtx.Unlock()
}

func (p *Package) ExportedType(name string) {
	p.mtx.Lock()
	p.ExportedTypes[name] = struct{}{}
	p.mtx.Unlock()
}

func (p *Package) ExportedVar(name string) {
	p.mtx.Lock()
	p.ExportedVariables[name] = struct{}{}
	p.mtx.Unlock()
}

func (p *Package) Print() {
	color.Green("========== %s (%s) ========", p.Name, p.Path)
	printPackage(p)
	printExportedItems(p)
}
func printPackage(pkg *Package) {
	color.Red("Imports: (%d)", len(pkg.DirectImportedPackages))
	cnt := 0
	for p := range pkg.DirectImportedPackages {
		cnt++
		color.Red("\t %d. %s", cnt, p)
	}
	color.HiBlue("Imported By: (%d)", len(pkg.ImportedByPackages))
	cnt = 0
	for p := range pkg.ImportedByPackages {
		cnt++
		color.HiBlue("\t %d. %s", cnt, p)
	}
}
func printExportedItems(pkg *Package) {
	color.HiMagenta("Exported Functions: (%d)", len(pkg.ExportedFunctions))
	cnt := 0
	for p := range pkg.ExportedFunctions {
		cnt++
		color.HiMagenta("\t %d. %s", cnt, p)
	}
	color.HiGreen("Exported Types: (%d)", len(pkg.ExportedTypes))
	cnt = 0
	for p := range pkg.ExportedTypes {
		cnt++
		color.HiGreen("\t %d. %s", cnt, p)
	}
	color.HiRed("Exported Variables: (%d)", len(pkg.ExportedVariables))
	cnt = 0
	for p := range pkg.ExportedVariables {
		cnt++
		color.HiRed("\t %d. %s", cnt, p)
	}
}
