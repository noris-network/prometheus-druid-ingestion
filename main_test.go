package main

import (
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractUniqueLabels(t *testing.T) {
	var testData = []struct {
		name     string
		in       model.Value
		expected LabelSet
		errFn    func(error) bool
	}{
		{
			name: "empty vector",
			in:   model.Vector{},
			expected: func() LabelSet {
				var ls LabelSet
				return ls
			}(),
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			actual, err := extractUniqueLabels(test.in)
			if test.errFn != nil {
				assert.Error(t, err, "expected error on %s", test.name)
				assert.True(t, test.errFn(err), "unexpected error behaviour with %s", test.name)
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}
