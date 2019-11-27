package main

import (
	"github.com/moredhel/lagg/lagg"
	"os"
)

func main() {
	p := lagg.NewDefaultStream()

	p.Start()
	defer p.Stop()

	p.ParseStream(os.Stdin)
}
