package main

import "os"

func main() {
	SelectAndRunTest(getTestDir(os.Args))
}
