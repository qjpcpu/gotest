// +build darwin

package main

import (
	"os/exec"
)

func WriteCmdHistory(cmd string) {
	copyCmd := exec.Command("pbcopy")
	in, err := copyCmd.StdinPipe()
	if err != nil {
		return
	}

	if err := copyCmd.Start(); err != nil {
		return
	}
	if _, err := in.Write([]byte(cmd)); err != nil {
		return
	}
	if err := in.Close(); err != nil {
		return
	}
	copyCmd.Wait()
}
