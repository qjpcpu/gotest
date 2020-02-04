package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/qjpcpu/common/debug"
)

func SelectSingleTest(dirname string, lastItem *Item) (name, fn string) {
	suites := LoadTestFiles(dirname)
	if suites.Size() == 0 {
		debug.Print("No tests found")
		return
	}
	suites = ReorderByHistory(suites, dirname, lastItem)
	if suites.Size() > 20 {
		suiteNames := suites.SuiteNames()
		_, name = debug.Select("Select test suite", suiteNames, func(s *debug.SelectWidget) {
			s.Size = 20
			s.IsVimMode = true
			s.HideSelected = true
			s.Searcher = func(input string, index int) bool {
				return strings.Contains(strings.ToLower(suiteNames[index]), strings.ToLower(input))
			}
		})
		if len(suites.SuiteFunctions(name)) > 0 {
			_, fn = debug.Select("Select test function", suites.SuiteFunctions(name), func(s *debug.SelectWidget) {
				s.Size = 20
				s.IsVimMode = true
				s.HideSelected = true
				s.StartInSearchMode = false
				fns := suites.SuiteFunctions(name)
				s.Searcher = func(input string, index int) bool {
					return strings.Contains(strings.ToLower(fns[index]), strings.ToLower(input))
				}
			})
		}
	} else {
		var list []string
		for _, n := range suites.SuiteNames() {
			fns := suites.SuiteFunctions(n)
			if len(fns) > 0 {
				for _, f := range fns {
					list = append(list, fmt.Sprintf("%s.%s", n, f))
				}
			} else {
				list = append(list, n)
			}
		}
		_, res := debug.Select("Select test function", list, func(s *debug.SelectWidget) {
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
		debug.AllowPanic(func() {
			arr := strings.Split(res, ".")
			name = arr[0]
			fn = arr[1]
		})
	}
	return
}

func buildTestCommand(dir string, name, fn string) string {
	format := "go test --run '^%s$' --testify.m '^%s$' --test.v"
	args := []interface{}{name, fn}
	if fn == "" {
		format = "go test --run '^%s$' --test.v"
		args = []interface{}{name}
	}

	wd, err := os.Getwd()
	debug.ShouldBeNil(err)
	wd, err = filepath.Abs(wd)
	debug.ShouldBeNil(err)
	dirAbs, err := filepath.Abs(dir)
	debug.ShouldBeNil(err)
	if wd != dirAbs {
		format = "cd '%s' && " + format
		args = append([]interface{}{dir}, args...)
	}
	debug.Print(format, args...)
	return fmt.Sprintf(format, args...)
}

func SelectAndRunTest(dir string) {
	item := History.Get(dir)
	name, fn := SelectSingleTest(dir, item)
	if len(name) == 0 {
		return
	}
	cmd := buildTestCommand(dir, name, fn)
	History.Append(Item{Dir: dir, Test: name, Module: fn})
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
