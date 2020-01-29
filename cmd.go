package main

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/qjpcpu/common/debug"
)

func SelectSingleTest(dirname string) string {
	suites := LoadTestFiles(dirname)
	var name, fn string
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
			s.StartInSearchMode = true
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
			s.StartInSearchMode = true
			s.Searcher = func(input string, index int) bool {
				return strings.Contains(strings.ToLower(list[index]), strings.ToLower(input))
			}
		})
		name, fn = strings.Split(res, ".")[0], strings.Split(res, ".")[1]
	}
	if name == "" {
		return ""
	}
	format := "go test --test.v --run '^%s$' --testify.m '^%s$'"
	args := []interface{}{name, fn}
	if fn == "" {
		format = "go test --test.v --run '^%s$'"
		args = []interface{}{name}
	}
	debug.Print(format, args...)
	cmd := fmt.Sprintf(format, args...)
	return cmd
}
