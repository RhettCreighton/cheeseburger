package coverage

import (
	"reflect"
	"testing"
)

func TestMergeLineRanges(t *testing.T) {
	tests := []struct {
		name   string
		input  []LineRange
		expect []LineRange
	}{
		{
			name:   "empty input",
			input:  []LineRange{},
			expect: []LineRange{},
		},
		{
			name:   "single range",
			input:  []LineRange{{Start: 5, End: 10}},
			expect: []LineRange{{Start: 5, End: 10}},
		},
		{
			name:   "non overlapping ranges",
			input:  []LineRange{{Start: 1, End: 2}, {Start: 4, End: 5}},
			expect: []LineRange{{Start: 1, End: 2}, {Start: 4, End: 5}},
		},
		{
			name:   "overlapping ranges",
			input:  []LineRange{{Start: 1, End: 3}, {Start: 2, End: 5}},
			expect: []LineRange{{Start: 1, End: 5}},
		},
		{
			name:   "adjacent ranges",
			input:  []LineRange{{Start: 1, End: 3}, {Start: 4, End: 6}},
			expect: []LineRange{{Start: 1, End: 6}},
		},
		{
			name:   "unsorted input with merging",
			input:  []LineRange{{Start: 10, End: 12}, {Start: 1, End: 4}, {Start: 5, End: 8}},
			expect: []LineRange{{Start: 1, End: 8}, {Start: 10, End: 12}},
		},
		{
			name: "complex merging scenarios",
			input: []LineRange{
				{Start: 5, End: 6},
				{Start: 1, End: 3},
				{Start: 2, End: 4},
				{Start: 8, End: 10},
				{Start: 11, End: 11}, // adjacent to previous
				{Start: 15, End: 20},
				{Start: 14, End: 16}, // overlapping with previous
			},
			expect: []LineRange{
				{Start: 1, End: 6},
				{Start: 8, End: 11},
				{Start: 14, End: 20},
			},
		},
		{
			name: "edge case: merging equals boundaries",
			input: []LineRange{
				{Start: 3, End: 5},
				{Start: 6, End: 8},  // note: adjacent because 6 == 5+1
				{Start: 9, End: 10}, // adjacent to previous because 9 == 8+1
			},
			expect: []LineRange{
				{Start: 3, End: 10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeLineRanges(tt.input)
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("mergeLineRanges() = %v, want %v", got, tt.expect)
			}
		})
	}
}
