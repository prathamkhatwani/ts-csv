package csv

import "testing"

func FuzzParse(f *testing.F) {
	// Add seed corpus
	f.Add("")
	f.Add("a,b,c\n1,2,3")
	f.Add("a,\"b,\",c\n1,2,3")
	f.Add("a,\"b\\\"\",c\n1,2,3")
	f.Add("a,\"b\n\",c\n1,2,3")

	f.Fuzz(func(t *testing.T, data string) {
		parser := NewParser()
		_, _ = parser.Parse(data)
	})
}
