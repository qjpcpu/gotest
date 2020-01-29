package main

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/qjpcpu/common/debug"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func execCommand(cmdstr string) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}
	binary, lookErr := exec.LookPath(shell)
	if lookErr != nil {
		panic(lookErr)
	}
	args := []string{binary, "-i", "-c", cmdstr}
	env := os.Environ()
	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}
}

func SelectSingleTest(dirname string) string {
	suites := LoadTestFiles(dirname)
	_, name := debug.Select("Select test suite", suites.SuiteNames(), func(s *promptui.Select) {
		s.Size = 20
		s.IsVimMode = true
		s.HideSelected = true
	})
	_, fn := debug.Select("Select test function", suites.SuiteFunctions(name), func(s *promptui.Select) {
		s.Size = 20
		s.IsVimMode = true
		s.HideSelected = true
		s.StartInSearchMode = true
		fns := suites.SuiteFunctions(name)
		s.Searcher = func(input string, index int) bool {
			return strings.Contains(strings.ToLower(fns[index]), strings.ToLower(input))
		}
	})
	debug.Print("go test --test.v --run %s --testify.m %s", name, fn)
	cmd := fmt.Sprintf(`go test --test.v --run %s --testify.m %s`, name, fn)
	return cmd
}
