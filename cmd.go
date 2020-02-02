package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/qjpcpu/common/debug"
)

func SelectSingleTest(dirname string) (name, fn string) {
	suites := LoadTestFiles(dirname)
	if suites.Size() == 0 {
		debug.Print("No tests found")
		return
	}
	if suites.Size() > 20 {
		_, name = debug.Select("Select test suite", suites.SuiteNames(), func(s *promptui.Select) {
			s.Size = 20
			s.IsVimMode = true
			s.HideSelected = true
		})
		_, fn = debug.Select("Select test function", suites.SuiteFunctions(name), func(s *promptui.Select) {
			s.Size = 20
			s.IsVimMode = true
			s.HideSelected = true
			s.StartInSearchMode = false
			fns := suites.SuiteFunctions(name)
			s.Searcher = func(input string, index int) bool {
				return strings.Contains(strings.ToLower(fns[index]), strings.ToLower(input))
			}
		})
	} else {
		var list []string
		for _, n := range suites.SuiteNames() {
			fns := suites.SuiteFunctions(n)
			for _, f := range fns {
				list = append(list, fmt.Sprintf("%s.%s", n, f))
			}
		}
		_, res := debug.Select("Select test function", list, func(s *promptui.Select) {
			s.Size = 20
			s.IsVimMode = true
			s.HideSelected = true
			s.StartInSearchMode = false
			s.Searcher = func(input string, index int) bool {
				return strings.Contains(strings.ToLower(list[index]), strings.ToLower(input))
			}
		})
		if res == "" {
			return
		}
		name, fn = strings.Split(res, ".")[0], strings.Split(res, ".")[1]
	}
	return
}

func buildTestCommand(dir string, name, fn string) string {
	format := "go test --test.v --run '^%s$' --testify.m '^%s$'"
	args := []interface{}{name, fn}
	if fn == "" {
		format = "go test --test.v --run '^%s$'"
		args = []interface{}{name}
	}

	wd, err := os.Getwd()
	debug.ShouldBeNil(err)
	wd, err = filepath.Abs(wd)
	debug.ShouldBeNil(err)
	dir, err = filepath.Abs(dir)
	debug.ShouldBeNil(err)
	if wd != dir {
		format = "cd '%s' && " + format
		args = append([]interface{}{dir}, args...)
	}
	debug.Print(format, args...)
	return fmt.Sprintf(format, args...)
}

func SelectAndRunTest(dir string) {
	name, fn := SelectSingleTest(dir)
	if len(name) == 0 {
		return
	}
	cmd := buildTestCommand(dir, name, fn)
	WriteCmdHistory(cmd)
	debug.Exec(cmd)
}

func getTestDir(args []string) string {
	const currentDir = "."
	var dir string
	if len(args) > 1 {
		dir = args[1]
	} else {
		dir = currentDir
	}
	return dir
}
