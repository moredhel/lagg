package main

import (
	"os"
	"time"

	"github.com/mkideal/cli"
	"github.com/moredhel/lagg/lagg"

)

type argT struct {
	cli.Helper
	WindowSize  int    `cli:"windowsize" usage:"How many lines to output" dft:"16"`
	MaxSize     int `cli:"maxsize" usage:"maximum number of logs to keep before we start compacting" dft:"1024"`
	RefreshInterval  int    `cli:"refreshinterval" usage:"refresh interval of the output (ms)" dft:"100"`
}

func main() {
	os.Exit(cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		refreshInterval := time.Millisecond * time.Duration(argv.RefreshInterval)
		p := lagg.NewStream(os.Stdin, argv.WindowSize, argv.MaxSize, refreshInterval)

		p.Start()
		defer p.Stop()

		p.ParseStream(os.Stdin)
		return nil
	}))
}
