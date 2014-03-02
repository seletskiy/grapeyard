package builder

import (
    "errors"
    "os"

    "go/ast"
    "go/token"
    "go/parser"

    "code.google.com/p/go.tools/go/types"
  _ "code.google.com/p/go.tools/go/gcimporter"
)

func ok(os.FileInfo) bool {
    return true;
}


func LoadPackage(dir string) (*types.Package, error) {
    fset := token.NewFileSet()

    packages, err := parser.ParseDir(fset, dir, ok, 0)
    if err != nil {
        return nil, err
    }

    var pkg *ast.Package

    if len(packages) == 0 {
        return nil, errors.New("no packages found in given directory")
    }

    for _, astPkg := range packages {
        pkg = astPkg
        break
    }

    fileset := make([]*ast.File, len(pkg.Files))
    idx := 0

    for _, file := range pkg.Files {
        fileset[idx] = file
        idx++
    }

    typesPkg, err := types.Check(dir, fset, fileset)
    if err != nil {
        return nil, err
    }

    return typesPkg, nil
}

func FindImplentations(i *types.Interface, pkg *types.Package) ([]string) {
    var names []string

    scope := pkg.Scope()
    allNames := scope.Names()

    for _, name := range allNames {
        obj := scope.Lookup(name)
        if typeName, ok := obj.(*types.TypeName); ok {
            if types.Implements(typeName.Type(), i) {
                names = append(names, typeName.Name())
            } else {
                println(typeName.Name(), "cannot be an ensurer")
                println(types.NewMethodSet(typeName.Type()).String())
            }
        }
    }

    return names
}

