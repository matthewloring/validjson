package main

import (
	"github.com/matthewloring/validjson"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(validjson.Analyzer)
}
