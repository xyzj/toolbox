package main

import (
	"io"
	"time"

	ginmiddleware "github.com/xyzj/toolbox/ginmiddle"
	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/proc"
)

func main() {
	// f := flag.String("f", "", "set input file")
	// flag.Parse()
	var proce = proc.StartRecord(&proc.RecordOpt{
		Logg:        logger.NewConsoleLogger(),
		DataTimeout: time.Hour * 24 * 366,
	})
	r := ginmiddleware.LiteEngine(io.Discard)
	r.GET("/", proce.GinHandler)
	r.Run(":8080")
}
