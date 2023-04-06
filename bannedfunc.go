package bannedfunc

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// BannedFunc is the linter.
type BannedFunc struct {
	funcs map[string]string
	pkg   *types.Package
	files []*ast.File
}

// Msg is a message.
type Msg struct {
	Pos  token.Pos
	Tips string
}

// NewLinter returns a new bannedfunc linter.
func NewLinter(bannedfuncs map[string]string, pkg *types.Package, files []*ast.File) *BannedFunc {
	return &BannedFunc{
		funcs: bannedfuncs,
		pkg:   pkg,
		files: files,
	}
}

// Run runs this linter and returns a slice of messages.
func (bf *BannedFunc) Run() []*Msg {
	var (
		confMap = bf.parseBannedFunc()
		usedMap = bf.usedImports(confMap)

		msgs []*Msg
	)
	for _, file := range bf.files {
		ast.Inspect(file, func(n ast.Node) bool {
			selector, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			m := usedMap[ident.Name]
			if m == nil {
				return true
			}
			tips, ok := m[selector.Sel.Name]
			if !ok {
				tips, ok = m["*"]
				if !ok {
					return true
				}
			}
			msgs = append(msgs, &Msg{
				Pos:  n.Pos(),
				Tips: tips,
			})
			return true
		})
	}
	return msgs
}

// parseBannedFunc parses the banned function configuration.
// return: map[import]map[func]tips
// example:
// 	{
// 		"ioutil": {
// 			"WriteFile": "As of Go 1.16, this function simply calls os.WriteFile.",
// 			"ReadFile":"As of Go 1.16, this function simply calls os.ReadFile.",
// 		},
// 		"github.com/example/banned": {
// 			"New": "This function is deprecated",
// 		},
// 	}
func (bf *BannedFunc) parseBannedFunc() map[string]map[string]string {
	confMap := make(map[string]map[string]string, len(bf.funcs))
	for f, tips := range bf.funcs {
		first, last := strings.Index(f, "("), strings.Index(f, ")")
		if first < 0 || last <= 0 || first > last || first+1 == last {
			continue
		}
		var (
			importName = f[first+1 : last]
			funcName   = f[last+2:]
		)
		if importName == "" || funcName == "" {
			continue
		}
		if conf, ok := confMap[importName]; ok {
			conf[funcName] = tips
		} else {
			confMap[importName] = map[string]string{funcName: tips}
		}
	}
	for imp, conf := range confMap {
		for funcname, tips := range conf {
			if funcname == "*" {
				confMap[imp] = map[string]string{"*": tips}
			}
		}
	}
	return confMap
}

// usedImports returns the used imports.
// param: confMap: map[import]map[func]tips
// return: map[import]map[func]tips
func (bf *BannedFunc) usedImports(confMap map[string]map[string]string) map[string]map[string]string {
	usedMap := make(map[string]map[string]string)
	for _, imp := range bf.pkg.Imports() {
		if conf, ok := confMap[imp.Path()]; ok {
			usedMap[imp.Name()] = conf
		}
	}
	return usedMap
}
