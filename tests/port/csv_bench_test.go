package csv_test

import (
	csv "github.com/prathamkhatwani/ts-csv/src"
	"testing"
)

func BenchmarkParseSmall(b *testing.B) {
	parser := csv.NewParser()
	text := "a,b,c\n1,2,3\n4,5,6"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(text)
	}
}

func BenchmarkParseMedium(b *testing.B) {
	parser := csv.NewParser()
	text := "a,b,c\n"
	for i := 0; i < 100; i++ {
		text += "1,2,3\n"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(text)
	}
}

func BenchmarkParseLarge(b *testing.B) {
	parser := csv.NewParser()
	text := "a,b,c\n"
	for i := 0; i < 1000; i++ {
		text += "1,2,3\n"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(text)
	}
}
