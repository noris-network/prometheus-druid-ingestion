/*
Copyright 2020 noris network AG

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ingestion

import (
	"testing"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

const unixTimestamp = 1583395744

func newMetric(labelName, labelValue string) model.Metric {
	out := make(map[model.LabelName]model.LabelValue, 1)
	out[model.LabelName(labelName)] = model.LabelValue(labelValue)
	return out
}

func newSample(metric model.Metric, val float64, ts int64) *model.Sample {
	return &model.Sample{
		Metric:    metric,
		Value:     model.SampleValue(val),
		Timestamp: model.Time(ts),
	}
}

func TestExtractUniqueLabels(t *testing.T) {
	var testData = []struct {
		name     string
		in       model.Value
		expected LabelSet
		errFn    func(error) bool
	}{
		{
			name:     "empty vector",
			in:       model.Vector{},
			expected: LabelSet{},
		},
		{
			name: "vector with single sample",
			in: model.Vector{
				newSample(newMetric("foo", "bar"), 0, unixTimestamp),
			},
			expected: LabelSet{"foo"},
		},
		{
			name: "label __name__ should be thrown out",
			in: model.Vector{
				newSample(newMetric("foo", "bar"), 0, unixTimestamp),
				newSample(newMetric("__name__", "up"), 0, unixTimestamp+1),
			},
			expected: LabelSet{"foo"},
		},
		{
			name: "label __name__ should be empty",
			in: model.Vector{
				newSample(newMetric("__name__", "up"), 0, unixTimestamp+1),
			},
			expected: LabelSet{},
		},
		{
			name: "vector with multiple samples with same label name",
			in: model.Vector{
				newSample(newMetric("foo", "bar"), 0, unixTimestamp),
				newSample(newMetric("foo", "baz"), 0, unixTimestamp+1),
			},
			expected: LabelSet{"foo"},
		},
		{
			name: "vector with multiple samples with different labels",
			in: model.Vector{
				newSample(newMetric("foo", "bar"), 0, unixTimestamp),
				newSample(newMetric("foo", "baz"), 0, unixTimestamp+1),
				newSample(newMetric("test", "test"), 0, unixTimestamp+2),
			},
			expected: LabelSet{"foo", "test"},
		},
		{
			name: "matrix as input",
			in:   model.Matrix{},
			errFn: func(err error) bool {
				return err != nil && err.Error() == "query result is not a Vector"
			},
		},
		{
			name: "string as input",
			in:   &model.String{},
			errFn: func(err error) bool {
				return err != nil && err.Error() == "query result is not a Vector"
			},
		},
		{
			name: "scalar as input",
			in:   &model.Scalar{},
			errFn: func(err error) bool {
				return err != nil && err.Error() == "query result is not a Vector"
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ExtractUniqueLabels(test.in)
			if test.errFn != nil {
				assert.Error(t, err)
				assert.True(t, test.errFn(err))
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestLabelSetToFields(t *testing.T) {
	var testData = []struct {
		name     string
		in       LabelSet
		expected FieldList
	}{
		{
			name:     "empty LabelSet",
			in:       LabelSet{},
			expected: FieldList{},
		},
		{
			name: "single label",
			in:   LabelSet{"foo"},
			expected: FieldList{
				Field{
					Type: "path",
					Name: "foo",
					Expr: "$.labels.foo",
				},
				Field{
					Type: "root",
					Name: "name",
					Expr: "name",
				},
				Field{
					Type: "root",
					Name: "value",
					Expr: "value",
				},
			},
		},
		{
			name: "multiple labels",
			in:   LabelSet{"foo", "bar"},
			expected: FieldList{
				Field{
					Type: "path",
					Name: "foo",
					Expr: "$.labels.foo",
				},
				Field{
					Type: "path",
					Name: "bar",
					Expr: "$.labels.bar",
				},
				Field{
					Type: "root",
					Name: "name",
					Expr: "name",
				},
				Field{
					Type: "root",
					Name: "value",
					Expr: "value",
				},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			actual := test.in.ToFieldList()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestLabelSetToDimensions(t *testing.T) {
	var testData = []struct {
		name     string
		input    LabelSet
		expected []string
	}{
		{
			name:     "empty LabelSet",
			input:    LabelSet{},
			expected: []string{},
		},
		{
			name:     "single label",
			input:    LabelSet{"foo"},
			expected: []string{"name", "foo"},
		},
		{
			name:     "multiple labels",
			input:    LabelSet{"foo", "bar"},
			expected: []string{"name", "foo", "bar"},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			actual := test.input.ToDimensions()
			assert.Equal(t, test.expected, actual)
		})
	}
}
