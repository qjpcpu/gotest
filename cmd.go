package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/qjpcpu/common.v2/assert"
	"github.com/qjpcpu/common.v2/cli"
	cfmt "github.com/qjpcpu/common.v2/fmt"
	"github.com/qjpcpu/common.v2/stringutil"
	"github.com/qjpcpu/common.v2/sys"
)

type GotestArgs struct {
	Dir     string
	File    string
	IsDebug bool
	Verbose bool
	Timeout string
}

func SelectSingleTest(dirname, file string, lastItem *Item) (name, fn string) {
	suites := LoadTestFiles(dirname, file)
	if suites.Size() == 0 {
		cfmt.Print("No tests found")
		return
	}
	suites = ReorderByHistory(suites, dirname, lastItem)
	if suites.Size() > 20 {
		suiteNames := suites.SuiteNames()
		if i := cli.SelectWithSearch("Select test suite", suiteNames); i != -1 {
			name = suiteNames[i]
		}

		if len(suites.SuiteFunctions(name)) > 0 {
			if i := cli.SelectWithSearch("Select test function", suites.SuiteFunctions(name)); i != -1 {
				fn = suites.SuiteFunctions(name)[i]
			}

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
		var res string
		if i := cli.SelectWithSearch("Select test function", list); i != -1 {
			res = list[i]
		}

		if res == "" {
			return
		}
		assert.AllowPanic(func() {
			arr := strings.Split(res, ".")
			name = arr[0]
			fn = arr[1]
		})
	}
	return
}

func buildTestCommand(dir string, name, fn string, targs GotestArgs) string {
	var exe string
	if targs.IsDebug {
		exe = `dlv test -- `
	} else {
		exe = `go test `
	}
	format := "--test.run '^%s$' --testify.m '^%s$'"
	args := []interface{}{name, fn}
	if fn == "" {
		format = "--test.run '^%s$'"
		args = []interface{}{name}
	}
	if targs.Verbose {
		format += " --test.v"
	}
	if !stringutil.IsBlankStr(targs.Timeout) {
		format += fmt.Sprintf(" --test.timeout %s", targs.Timeout)
	}
	format = exe + format

	wd, err := os.Getwd()
	assert.ShouldBeNil(err)
	wd, err = filepath.Abs(wd)
	assert.ShouldBeNil(err)
	dirAbs, err := filepath.Abs(dir)
	assert.ShouldBeNil(err)
	if wd != dirAbs {
		format = "cd '%s' && " + format
		if strings.HasPrefix(dir, wd) && len(wd) > 0 {
			dir = strings.TrimPrefix(dir, wd)
			if strings.HasPrefix(dir, "/") {
				dir = strings.TrimPrefix(dir, "/")
			}
		}
		args = append([]interface{}{dir}, args...)
	}
	cfmt.Print(format, args...)
	return fmt.Sprintf(format, args...)
}

func SelectAndRunTest(args GotestArgs) {
	item := History.Get(args.Dir)
	name, fn := SelectSingleTest(args.Dir, args.File, item)
	if len(name) == 0 {
		return
	}
	cmd := buildTestCommand(args.Dir, name, fn, args)
	History.Append(Item{Dir: args.Dir, Test: name, Module: fn})
	sys.Exec(cmd)
}

func getTestArgs(args []string) (targs GotestArgs) {
	/* has verbose */
	if targs.Verbose = stringutil.ContainString(args, "-v"); targs.Verbose {
		args = stringutil.RemoveString(args, "-v")
	}
	/* timeout */
	if stringutil.ContainString(args, "-timeout") {
		i := 0
		for ; i < len(args); i++ {
			if args[i] == "-timeout" {
				targs.Timeout = args[i+1]
				break
			}
		}
		for j := i + 2; j < len(args); j++ {
			args[j-2] = args[j]
		}
		args = args[:len(args)-2]
	}

	if len(args) > 1 && args[1] == `debug` {
		targs.IsDebug = true
		args = args[1:]
	}
	const currentDir = "."
	if len(args) > 1 {
		targs.Dir = args[1]
		fi, err := os.Stat(targs.Dir)
		assert.ShouldBeNil(err)
		if !fi.IsDir() {
			targs.File = filepath.Base(targs.Dir)
			targs.Dir = filepath.Dir(targs.Dir)
		}
	} else {
		targs.Dir = currentDir
	}

	return
}
