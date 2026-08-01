package fuzz_test

import (
	csv "github.com/prathamkhatwani/ts-csv/src"
	"testing"
)

func FuzzParse(f *testing.F) {
	// Add seed corpus
	f.Add("")
	f.Add("a,b,c\n1,2,3")
	f.Add("a,\"b,\",c\n1,2,3")
	f.Add("a,\"b\\\"\",c\n1,2,3")
	f.Add("a,\"b\n\",c\n1,2,3")

	f.Fuzz(func(t *testing.T, data string) {
		parser := csv.NewParser()
		_, _ = parser.Parse(data)
	})
}
