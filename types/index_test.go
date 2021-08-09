package types

import (
	"testing"
)

type Event struct {
	Name        string `auditdb:"index" json:"name"`
	Location    string `auditdb:"index" json:"location"`
	CamelCase   string `auditdb:"index" json:"camel_case"`
	WithoutJson string `auditdb:"index"`
	NonIndex    string `json:"nonIndex"`
	unexported  string `auditdb:"index" json:"unexported"`
	TS          int64  `json:"-"`
}

func TestGetIndexes(t *testing.T) {
	indexes, rest := getIndexes(Event{
		Name:        "name",
		Location:    "location",
		CamelCase:   "camel_case",
		NonIndex:    "nonIndex",
		WithoutJson: "withoutJson",
	})

	if len(indexes) != 4 {
		t.Fatalf("Exected 4 indexes, got %d", len(indexes))
	}

	if len(rest) != 1 {
		t.Fatalf("Exected 1 rest, got %d", len(rest))
	}

	for name, val := range map[string]string{
		"name":        "name",
		"location":    "location",
		"camel_case":  "camel_case",
		"withoutJson": "withoutJson",
	} {
		if indexes[name] != val {
			t.Fatalf("Indexes don't match! Expected %v, got %v", val, indexes[name])
		}
	}
}

func TestLowerInitial(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{
			str:      "Name",
			expected: "name",
		},
		{
			str:      "NAME",
			expected: "nAME",
		},
	}
	for i, test := range tests {
		res := lowerInitial(test.str)
		if test.expected != res {
			t.Fatalf("test %d: Expected %s, got %s", i, test.expected, res)
		}
	}
}
