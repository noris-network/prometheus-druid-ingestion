package ingestion

import (
	"fmt"
	"github.com/prometheus/common/model"
)

// LabelSet is a unique set of Prometheus labels.
type LabelSet []string

// ToFieldList converts a LabelSet to a FieldList
func (labels LabelSet) ToFieldList() FieldList {
	fields := FieldList{}

	if len(labels) == 0 {
		return fields
	}

	for _, f := range labels {
		fields = append(fields, Field{
			Type: "path",
			Name: f,
			Expr: fmt.Sprintf("$.labels.%s", f),
		})
	}

	// These fields are added by prometheus-kafka-adapter
	fields = append(fields, Field{
		Type: "root",
		Name: "name",
		Expr: "name",
	})
	fields = append(fields, Field{
		Type: "root",
		Name: "value",
		Expr: "value",
	})

	return fields
}

// ToDimensions converts a LabelSet to a slice of strings that can be used
// for Druids Dimensions. It also adds the dimension 'name' to the slice.
func (labels LabelSet) ToDimensions() []string {
	if len(labels) == 0 {
		return []string{}
	}
	dimensions := make([]string, len(labels)+1)
	dimensions[0] = "name"
	for i := 1; i <= len(labels); i++ {
		dimensions[i] = labels[i-1]
	}
	return dimensions
}

// ExtractUniqueLabels extracts unique labels from a Prometheus query  result.
func ExtractUniqueLabels(result model.Value) (LabelSet, error) {
	vec, ok := result.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("query result is not a Vector")
	}

	labelSeen := make(map[string]bool)
	labels := LabelSet{}
	for _, m := range vec {
		for k := range m.Metric {
			key := string(k)
			// __name__ is thrown out
			if key == "__name__" {
				continue
			}
			// If we haven't seen this label before, add it to the list.
			if _, ok := labelSeen[key]; !ok {
				labels = append(labels, key)
				labelSeen[key] = true
			}
		}
	}
	return labels, nil
}
