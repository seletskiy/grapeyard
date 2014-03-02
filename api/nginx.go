package api

type Nginx struct{}


func (n Nginx) Ensure(cfg map[string]string) error {
    // install
    var p *Package

    p.install("nginx")
    return nil
}
