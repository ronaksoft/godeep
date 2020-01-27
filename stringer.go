package godeep

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"
)

/*
   Creation Time: 2020 - Jan - 27
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func astExpr(x ast.Expr) string {
	switch xx := x.(type) {
	case *ast.Ident:
		return astIdent(xx)
	case *ast.Ellipsis:
		return astEllipsis(xx)
	case *ast.ArrayType:
		return astArray(xx)
	case *ast.SelectorExpr:
		return astSelectorExpr(xx)
	case *ast.StarExpr:
		return astStarExpr(xx)
	case *ast.MapType:
		return astMapType(xx)
	case *ast.InterfaceType:
		return astInterfaceType(xx)
	case *ast.FuncType:
		return fmt.Sprintf("func %s", astFuncType(xx))
	default:
		return reflect.TypeOf(xx).String()
	}
}
func astStarExpr(x *ast.StarExpr) string {
	sb := strings.Builder{}
	sb.WriteRune('*')
	sb.WriteString(astExpr(x.X))
	return sb.String()
}
func astMapType(x *ast.MapType) string {
	sb := strings.Builder{}
	sb.WriteString("map[")
	sb.WriteString(astExpr(x.Key))
	sb.WriteString("]")
	sb.WriteString(astExpr(x.Value))
	return sb.String()
}
func astInterfaceType(x *ast.InterfaceType) string {
	return "interface{}"
}
func astFuncType(x *ast.FuncType) string {
	sb := strings.Builder{}
	sb.WriteRune('(')
	for idx, p := range x.Params.List {
		for idx, n := range p.Names {
			sb.WriteString(n.Name)
			if idx < len(p.Names)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteRune(' ')
		sb.WriteString(astExpr(p.Type))
		if idx < len(x.Params.List)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteRune(')')
	return sb.String()
}
func astSelectorExpr(x *ast.SelectorExpr) string {
	sb := strings.Builder{}
	sb.WriteString(astExpr(x.X))
	sb.WriteString(astIdent(x.Sel))
	return sb.String()
}
func astIdent(x *ast.Ident) string {
	return x.Name
}
func astEllipsis(x *ast.Ellipsis) string {
	sb := strings.Builder{}
	sb.WriteString("...")
	sb.WriteString(astExpr(x.Elt))
	return sb.String()
}
func astArray(x *ast.ArrayType) string {
	sb := strings.Builder{}
	sb.WriteString("[]")
	switch xx := x.Elt.(type) {
	case *ast.Ident:
		sb.WriteString(astIdent(xx))
	default:
		sb.WriteString(reflect.TypeOf(x.Elt).String())
	}
	return sb.String()
}
