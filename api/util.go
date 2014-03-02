package api

type Cron int

var Cronie Cron
var Filie File

func init() {
    Cronie = Cron(42)
    Filie = File("42")
}
