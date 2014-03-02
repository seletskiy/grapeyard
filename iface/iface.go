package iface


type Ensurer interface {
    Ensure(map[string]string) error
}
