package api


type Package string

func (p Package) Ensure(map[string]string) error {
    return nil
}


func (p Package) Install() error {
    return nil
}
