package main

import (
	"log"

	"github.com/jakecoffman/eng"
	"github.com/jakecoffman/fam"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	eng.Run(&fam.Game{})
}
