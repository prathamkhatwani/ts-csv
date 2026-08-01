package csv

import (
	"reflect"
	"regexp"
	"testing"
)

func toJson(keys []string, row []string) map[string]string {
	obj := make(map[string]string)
	for i, key := range keys {
		if i < len(row) {
			obj[key] = row[i]
		} else {
			obj[key] = ""
		}
	}
	return obj
}

func TestMetadata(t *testing.T) {
	if LibName != "@gregoranders/csv" {
		t.Errorf("expected LibName to be @gregoranders/csv, got %s", LibName)
	}

	if LibVersion != "0.0.13" {
		t.Errorf("expected LibVersion to be 0.0.13, got %s", LibVersion)
	}

	matched, err := regexp.MatchString("https", LibURL)
	if err != nil || !matched {
		t.Errorf("expected LibURL to match https, got %s", LibURL)
	}
}

func TestParserValidDefault(t *testing.T) {
	tests := []struct {
		text     string
		expected []Row
	}{
		{
			text:     "",
			expected: []Row{},
		},
		{
			text: "a,b,c\n1,2,3",
			expected: []Row{
				{"a", "b", "c"},
				{"1", "2", "3"},
			},
		},
		{
			text: "a,\"b,\",c\n1,2,3",
			expected: []Row{
				{"a", "b,", "c"},
				{"1", "2", "3"},
			},
		},
		{
			text: "a,\"b\\\"\",c\n1,2,3",
			expected: []Row{
				{"a", "b\"", "c"},
				{"1", "2", "3"},
			},
		},
		{
			text: "a,\"b\n\",c\n1,2,3",
			expected: []Row{
				{"a", "b\n", "c"},
				{"1", "2", "3"},
			},
		},
		{
			text: "a,b,c\n1,2,3\n",
			expected: []Row{
				{"a", "b", "c"},
				{"1", "2", "3"},
			},
		},
		{
			text: "a,b,c\n1,2,",
			expected: []Row{
				{"a", "b", "c"},
				{"1", "2", ""},
			},
		},
		{
			text: "\"a\",\"b\",\"c\"\n\"1\",\"2\",\"\"",
			expected: []Row{
				{"a", "b", "c"},
				{"1", "2", ""},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.text, func(t *testing.T) {
			parser := NewParser()
			rows, err := parser.Parse(tc.text)
			if err != nil {
				t.Fatalf("unexpected error parsing %q: %v", tc.text, err)
			}

			if !reflect.DeepEqual(rows, tc.expected) {
				t.Errorf("Parse() = %v, expected %v", rows, tc.expected)
			}

			if !reflect.DeepEqual(parser.Rows(), tc.expected) {
				t.Errorf("Rows() = %v, expected %v", parser.Rows(), tc.expected)
			}

			var expectedJSON []map[string]string
			if len(tc.expected) > 0 {
				expectedJSON = []map[string]string{toJson(tc.expected[0], tc.expected[1])}
			} else {
				expectedJSON = []map[string]string{}
			}

			jsonResult := parser.JSON()
			if !reflect.DeepEqual(jsonResult, expectedJSON) {
				t.Errorf("JSON() = %v, expected %v", jsonResult, expectedJSON)
			}
		})
	}
}

func TestParserInvalidDefault(t *testing.T) {
	tests := []struct {
		text        string
		expectedErr string
	}{
		{
			text:        "a,\"b\"\",c\n1,2,3",
			expectedErr: "Invalid CSV at 0:5",
		},
		{
			text:        "a,\"b\",c\n1,\"2\"\",",
			expectedErr: "Invalid CSV at 1:5",
		},
		{
			text:        "a,\"b,c\n1,2,",
			expectedErr: "Invalid CSV at 0:2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.text, func(t *testing.T) {
			parser := NewParser()
			_, err := parser.Parse(tc.text)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tc.expectedErr {
				t.Errorf("expected error %q, got %q", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestParserCustomConfigValid(t *testing.T) {
	tests := []struct {
		text     string
		expected []Row
	}{
		{
			text:     "",
			expected: []Row{},
		},
		{
			text: "a;b;c\t1;2;3",
			expected: []Row{
				{"a", "b", "c"},
				{"1", "2", "3"},
			},
		},
		{
			text: "a;'b\n';c\t1;2;3",
			expected: []Row{
				{"a", "b\n", "c"},
				{"1", "2", "3"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.text, func(t *testing.T) {
			parser := NewParser(Configuration{
				FieldSeparator: ";",
				Quote:          "'",
				LineSeparator:  "\t",
			})
			rows, err := parser.Parse(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(rows, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, rows)
			}

			if !reflect.DeepEqual(parser.Rows(), tc.expected) {
				t.Errorf("expected Rows() = %v, got %v", tc.expected, parser.Rows())
			}

			var expectedJSON []map[string]string
			if len(tc.expected) > 0 {
				expectedJSON = []map[string]string{toJson(tc.expected[0], tc.expected[1])}
			} else {
				expectedJSON = []map[string]string{}
			}

			jsonResult := parser.JSON()
			if !reflect.DeepEqual(jsonResult, expectedJSON) {
				t.Errorf("expected JSON() = %v, got %v", expectedJSON, jsonResult)
			}
		})
	}
}

func TestParserCustomConfigInvalid(t *testing.T) {
	tests := []struct {
		text        string
		config      Configuration
		expectedErr string
	}{
		{
			text: "a;\"b\"\";c\n1,2,3",
			config: Configuration{
				FieldSeparator: ";",
			},
			expectedErr: "Invalid CSV at 0:5",
		},
	}

	for _, tc := range tests {
		t.Run(tc.text, func(t *testing.T) {
			parser := NewParser(tc.config)
			_, err := parser.Parse(tc.text)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tc.expectedErr {
				t.Errorf("expected error %q, got %q", tc.expectedErr, err.Error())
			}
		})
	}
}
