package main

import (
	"github.com/qjpcpu/common/debug"
)

func main() {
	if cmd := SelectSingleTest("."); cmd != "" {
		debug.Exec(cmd)
	}
}
