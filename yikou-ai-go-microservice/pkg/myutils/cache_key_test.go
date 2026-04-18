package myutils

import (
	"testing"
)

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantLen int
	}{
		{
			name:    "nil input",
			input:   nil,
			wantLen: 32,
		},
		{
			name:    "string input",
			input:   "test",
			wantLen: 32,
		},
		{
			name:    "int input",
			input:   123,
			wantLen: 32,
		},
		{
			name:    "struct input",
			input:   struct{ Name string }{Name: "test"},
			wantLen: 32,
		},
		{
			name:    "map input",
			input:   map[string]string{"key": "value"},
			wantLen: 32,
		},
		{
			name:    "slice input",
			input:   []string{"a", "b", "c"},
			wantLen: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCacheKey(tt.input)
			if len(got) != tt.wantLen {
				t.Errorf("GenerateCacheKey() length = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestGenerateCacheKeyConsistency(t *testing.T) {
	input := map[string]string{"key": "value"}

	key1 := GenerateCacheKey(input)
	key2 := GenerateCacheKey(input)

	if key1 != key2 {
		t.Errorf("GenerateCacheKey() not consistent, got %v and %v", key1, key2)
	}
}

func TestGenerateCacheKeyDifferent(t *testing.T) {
	input1 := map[string]string{"key": "value1"}
	input2 := map[string]string{"key": "value2"}

	key1 := GenerateCacheKey(input1)
	key2 := GenerateCacheKey(input2)

	if key1 == key2 {
		t.Errorf("GenerateCacheKey() should produce different keys for different inputs")
	}
}
