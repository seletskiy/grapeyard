package main

import (
    "os"

    "go/ast"
    "go/token"
    "go/parser"

    "code.google.com/p/go.tools/go/types"
  _ "code.google.com/p/go.tools/go/gcimporter"
)

func ok(os.FileInfo) bool { return true; }

func main() {

    // core file

    fset := token.NewFileSet()

    packages, err := parser.ParseDir(fset, "../", ok, 0)
    if err != nil {
        panic(err)
    }

    core := packages["grapeyard"]
    fileset := make([]*ast.File, len(core.Files))

    idx := 0
    for _, file := range core.Files {
        fileset[idx] = file
        idx++
    }

    corePkg, err := types.Check("../", fset, fileset)
    if err != nil {
        panic(err)
    }


    fset = token.NewFileSet()
    packages, err = parser.ParseDir(fset, "../api", ok, 0)

    api := packages["api"]
    fileset = make([]*ast.File, len(api.Files))

    idx = 0
    for _, file := range api.Files {
        fileset[idx] = file
        idx++
    }

    apiPkg, err := types.Check("../api", fset, fileset)
    if err != nil {
        panic(err)
    }


    ensurer := corePkg.Scope().Lookup("Ensurer").Type().Underlying().(*types.Interface)
    names := apiPkg.Scope().Names()

    for _, name := range names {
        obj := apiPkg.Scope().Lookup(name)

        if types.Implements(obj.Type(), ensurer) {
            println(obj.Name(), "is an ensurer")
        } else {
            println(obj.Name(), "is not an ensurer")
        }
    }
}
