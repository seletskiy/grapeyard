package grapeyard


type Ensurer interface {
    Ensure(map[string]string) error
}
