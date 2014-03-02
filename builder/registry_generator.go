package builder

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"

    "code.google.com/p/go.tools/go/types"
    _ "code.google.com/p/go.tools/go/gcimporter"
)


const REGISTRY_HEADER = `type GenFunc func() iface.Ensurer
var Registry = make(map[string]GenFunc)
func init() {`

const REGISTRY_ITEM_BODY = `    Registry["%s"] = func() iface.Ensurer { return new(%s.%s); }`

const REGISTRY_FOOTER = `}`

func MakeRegistry(root string) (map[string][]string, error) {
    registry := make(map[string][]string)
    ifacePkg, err := LoadPackage(filepath.Join(root, "iface"))
    if err != nil {
        return nil, err
    }

    ensurer := ifacePkg.Scope().Lookup("Ensurer").Type().Underlying().(*types.Interface)

    // add stdlib api
    err = traverseDir(registry, "api", root, ensurer)
    if err != nil {
        return nil, err
    }

    err = traverseDir(registry, "user", root, ensurer)
    if err != nil {
        return nil, err
    }

    return registry, nil
}


func WriteRegistry(registry map[string][]string, dir string, baseUrl string) error {
    path := filepath.Join(dir, "registry")
    if _, err := os.Stat(path); os.IsNotExist(err) {
        err = os.MkdirAll(path, 0777)
    }

    regFile, err := os.Create(filepath.Join(path, "registry.go"))
    if err != nil {
        return err
    }

    // add base repo
    var imports = []string{filepath.Join(baseUrl, "iface")}

    for subpkg := range registry {
        imports = append(imports, filepath.Join(baseUrl, subpkg))
    }

    // write header and imports
    regFile.WriteString("package registry\n\n")

    regFile.WriteString("import (\n")

    for _, imp := range imports {
        fmt.Fprintf(regFile, "    \"%s\"\n", imp)
    }

    regFile.WriteString(")\n\n")

    // write registry

    regFile.WriteString(REGISTRY_HEADER)
    regFile.WriteString("\n")

    for subpkg, names := range registry {
        for _, name := range names {
            fmt.Fprintf(regFile, REGISTRY_ITEM_BODY, subpkg + "." + name, filepath.Base(subpkg),
                name)
            regFile.WriteString("\n")
        }
        regFile.WriteString("\n")
    }

    regFile.WriteString(REGISTRY_FOOTER)
    regFile.WriteString("\n")

    return nil
}


func traverseDir(registry map[string][]string, dir, base string, ensurer *types.Interface) error {
    fullPath := filepath.Join(base, dir)
    fileInfos, err := ioutil.ReadDir(fullPath)
    if err != nil {
        return err
    }

    for _, file := range fileInfos {
        if ! file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
            pkg, err := LoadPackage(fullPath)
            if err != nil {
                return err
            }

            names := FindImplentations(ensurer, pkg)

            // add type names into registry
            if len(names) != 0 {
                registry[dir] = names
            }
            break
        }
    }

    for _, fileInfo := range fileInfos {
        if fileInfo.IsDir() {
            // recur
            err = traverseDir(registry, filepath.Join(dir, fileInfo.Name()), base, ensurer)
            if err != nil {
                return err
            }
        }
    }

    return nil
}
