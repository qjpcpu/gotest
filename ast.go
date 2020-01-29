package main

import (
	"github.com/qjpcpu/common/debug"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

func LoadTestFiles(dirname string) FileTestSuite {
	fileList, err := ioutil.ReadDir(dirname)
	debug.ShouldBeNil(err)

	var files []string
	for _, file := range fileList {
		if strings.HasSuffix(file.Name(), "_test.go") {
			filename := filepath.Join(dirname, file.Name())
			filename, err := filepath.Abs(filename)
			debug.ShouldBeNil(err)
			files = append(files, filename)
		}
	}

	suite := newSuite()
	for _, file := range files {
		suite = suite.merge(ParseTestSuiteFile(file))
	}
	return suite
}

func ParseTestSuiteFile(filename string) FileTestSuite {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	debug.ShouldBeNil(err)

	typeToMainFunc := make(map[string]string)
	typeMethods := make(map[string][]string)
	for _, decl := range f.Decls {
		/* function */
		if declFn, ok := decl.(*ast.FuncDecl); ok {
			safeRun(func() {
				declFn := decl.(*ast.FuncDecl)
				name := declFn.Name.Name
				if !strings.HasPrefix(name, "Test") {
					panic("bad type")
				}
				if declFn.Type.Params.List[0].Type.(*ast.StarExpr).X.(*ast.SelectorExpr).X.(*ast.Ident).Name != "testing" {
					panic("bad test")
				}
				if declFn.Type.Params.List[0].Type.(*ast.StarExpr).X.(*ast.SelectorExpr).Sel.Name != "T" {
					panic("bad test")
				}
				tname := declFn.Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr).Args[1].(*ast.UnaryExpr).X.(*ast.CompositeLit).Type.(*ast.Ident).Name
				typeToMainFunc[tname] = name
			})
			safeRun(func() {
				name := declFn.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
				if !strings.HasPrefix(declFn.Name.Name, "Test") {
					panic("bad func")
				}
				typeMethods[name] = append(typeMethods[name], declFn.Name.Name)
			})
		}
	}
	ret := make(map[string][]string)
	for tp, methods := range typeMethods {
		if fn, ok := typeToMainFunc[tp]; ok {
			ret[fn] = methods
		}
	}
	return makeSuite(ret)
}

func safeRun(f func()) (ok bool) {
	defer func() {
		recover()
		ok = false
	}()
	f()
	ok = true
	return
}

type FileTestSuite struct {
	testFunctions map[string][]string
	testNames     []string
}

func newSuite() FileTestSuite {
	return FileTestSuite{
		testFunctions: make(map[string][]string),
	}
}

func makeSuite(ret map[string][]string) FileTestSuite {
	s := FileTestSuite{
		testFunctions: ret,
	}
	for n := range ret {
		s.testNames = append(s.testNames, n)
	}
	sort.Strings(s.testNames)
	return s
}

func (s FileTestSuite) merge(s1 FileTestSuite) FileTestSuite {
	s2 := FileTestSuite{
		testFunctions: make(map[string][]string),
	}
	for k, v := range s.testFunctions {
		s2.testFunctions[k] = append(s2.testFunctions[k], v...)
		s2.testNames = append(s2.testNames, k)
	}
	for k, v := range s1.testFunctions {
		s2.testFunctions[k] = append(s2.testFunctions[k], v...)
		s2.testNames = append(s2.testNames, k)
	}

	sort.Strings(s2.testNames)
	return s2
}

func (s FileTestSuite) SuiteNames() []string {
	return s.testNames
}

func (s FileTestSuite) SuiteFunctions(name string) []string {
	return s.testFunctions[name]
}
