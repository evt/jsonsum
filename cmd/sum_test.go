package main

import (
	"math/big"
	"testing"
)

func TestFindAndSumNumbers(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected *big.Int
	}{
		{
			name:     "Empty Input",
			input:    "",
			expected: big.NewInt(0),
		},
		{
			name:     "Integer Input",
			input:    123,
			expected: big.NewInt(123),
		},
		{
			name:     "Float Input",
			input:    3.14,
			expected: big.NewInt(3),
		},
		{
			name:     "String Input",
			input:    "-1 1 2 3.14 0x4",
			expected: big.NewInt(9),
		},
		{
			name: "Array Input",
			input: []any{
				1,
				2,
				3,
			},
			expected: big.NewInt(6),
		},
		{
			name: "Nested Array Input",
			input: []any{
				1,
				[]any{
					2,
					3,
				},
			},
			expected: big.NewInt(6),
		},
		{
			name: "Map Input",
			input: map[string]any{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			expected: big.NewInt(6),
		},
		{
			name: "Nested Map Input",
			input: map[string]any{
				"a": 1,
				"b": map[string]any{
					"c": 2,
					"d": 3,
				},
			},
			expected: big.NewInt(6),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := findAndSumNumbers(tc.input)

			if result.Cmp(tc.expected) != 0 {
				t.Errorf("Result: %v, Expected: %v", result, tc.expected)
			}
		})
	}
}
