package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/qjpcpu/common.v2/assert"
)

func LoadTestFiles(dirname, gofile string) FileTestSuite {
	fileList, err := ioutil.ReadDir(dirname)
	assert.ShouldBeNil(err)

	var files []string
	for _, file := range fileList {
		if gofile != "" && gofile != file.Name() {
			continue
		}
		if strings.HasSuffix(file.Name(), "_test.go") {
			filename := filepath.Join(dirname, file.Name())
			filename, err := filepath.Abs(filename)
			assert.ShouldBeNil(err)
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
	assert.ShouldBeNil(err)

	typeToMainFunc := make(map[string]string)
	typeMethods := make(map[string][]string)
	simpleGoTest := make(map[string]bool)
	for _, decl := range f.Decls {
		/* function */
		if declFn, ok := decl.(*ast.FuncDecl); ok {
			assert.AllowPanic(func() {
				declFn := decl.(*ast.FuncDecl)
				name := declFn.Name.Name
				assert.ShouldBeTrue(strings.HasPrefix(name, "Test"))

				assert.ShouldEqual(declFn.Type.Params.List[0].Type.(*ast.StarExpr).X.(*ast.SelectorExpr).X.(*ast.Ident).Name, "testing")

				assert.ShouldEqual(declFn.Type.Params.List[0].Type.(*ast.StarExpr).X.(*ast.SelectorExpr).Sel.Name, "T")

				if isSimpleTest := assert.AllowPanic(func() {
					callExpr := declFn.Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr)
					assert.ShouldEqual(callExpr.Fun.(*ast.SelectorExpr).Sel.Name, "Run")
					assert.ShouldSuccessAtLeastOne(
						func() {
							tname := callExpr.Args[1].(*ast.UnaryExpr).X.(*ast.CompositeLit).Type.(*ast.Ident).Name
							typeToMainFunc[tname] = name
						},
						func() {
							tname := callExpr.Args[1].(*ast.CallExpr).Args[0].(*ast.Ident).Name
							typeToMainFunc[tname] = name
						},
						func() {
							tname := callExpr.Args[1].(*ast.CompositeLit).Type.(*ast.Ident).Name
							typeToMainFunc[tname] = name
						},
					)
				}); isSimpleTest {
					simpleGoTest[name] = true
				}
			})
			assert.AllowPanic(func() {
				assert.ShouldSuccessAtLeastOne(
					func() {
						name := declFn.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
						assert.ShouldBeTrue(strings.HasPrefix(declFn.Name.Name, "Test"))
						typeMethods[name] = append(typeMethods[name], declFn.Name.Name)
					},
					func() {
						name := declFn.Recv.List[0].Type.(*ast.Ident).Name
						assert.ShouldBeTrue(strings.HasPrefix(declFn.Name.Name, "Test"))
						typeMethods[name] = append(typeMethods[name], declFn.Name.Name)
					},
				)

			})
		}
	}
	ret := make(map[string][]string)
	for fn := range simpleGoTest {
		ret[fn] = []string{}
	}
	for tp, methods := range typeMethods {
		if fn, ok := typeToMainFunc[tp]; ok {
			ret[fn] = methods
		}
	}
	return makeSuite(ret)
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

func (s FileTestSuite) Size() int {
	var total int
	for _, v := range s.testFunctions {
		if len(v) > 0 {
			total += len(v)
		} else {
			total++
		}
	}
	return total
}

func (s FileTestSuite) SetTop(name, fn string) FileTestSuite {
	for i, n := range s.testNames {
		if n == name {
			s.testNames[0], s.testNames[i] = s.testNames[i], s.testNames[0]
			fns := s.testFunctions[n]
			for j, f := range fns {
				if fn != "" && f == fn {
					fns[0], fns[j] = fns[j], fns[0]
					break
				}
			}
			s.testFunctions[n] = fns
			break
		}
	}
	return s
}

func ReorderByHistory(s FileTestSuite, dir string, item *Item) FileTestSuite {
	if item != nil {
		dirAbs, err := filepath.Abs(dir)
		assert.ShouldBeNil(err)
		hAbs, err := filepath.Abs(item.Dir)
		assert.ShouldBeNil(err)
		if hAbs == dirAbs {
			s = s.SetTop(item.Test, item.Module)
		}
	}
	return s
}
