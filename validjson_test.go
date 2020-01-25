package validjson_test

import (
	"testing"

	"github.com/matthewloring/validjson"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, validjson.Analyzer, "a")
}
