package builder

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)


func StripSources(dir string) error {
    infos, err := ioutil.ReadDir(dir)
    if err != nil {
        return err
    }

    for _, info := range infos {
        if ! info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
            err = os.Remove(filepath.Join(dir, info.Name()))
            if err != nil {
                return err
            }
        } else if info.IsDir() {
            err = StripSources(filepath.Join(dir, info.Name()))
            if err != nil {
                return err
            }
        }
    }

    return nil
}
