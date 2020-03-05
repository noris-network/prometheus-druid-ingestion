package modify

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	jsonComplete = `{
    "type": "kafka",
    "dataSchema": {
        "dataSource": "test",
        "parser": {
            "type": "string",
            "parseSpec": {
                "format": "json",
                "timestampSpec": {
                    "column": "timestamp",
                    "format": "iso"
                },
                "flattenSpec": {
                    "fields": [
                        {
                            "type": "path",
                            "name": "instance",
                            "expr": "$.labels.instance"
                        },
                        {
                            "type": "path",
                            "name": "job",
                            "expr": "$.labels.job"
                        },
                        {
                            "type": "root",
                            "name": "name",
                            "expr": "name"
                        },
                        {
                            "type": "root",
                            "name": "value",
                            "expr": "value"
                        }
                    ]
                },
                "dimensionsSpec": {
                    "dimensions": [
                        "name",
                        "instance",
                        "job"
                    ]
                }
            }
        },
        "metricsSpec": [
            {
                "name": "count",
                "type": "count"
            },
            {
                "name": "value",
                "type": "doubleMax",
                "fieldName": "value"
            }
        ],
        "granularitySpec": {
            "type": "uniform",
            "segmentGranularity": "HOUR",
            "queryGranularity": "MINUTE"
        }
    },
    "ioConfig": {
        "topic": "test",
        "consumerProperties": {
            "bootstrap.servers": "test"
        },
        "taskDuration": "PT10M",
        "useEarliestOffset": true
    }
}`
)

func defaultKafkaIngestionSpec() *KafkaIngestionSpec {
	return &KafkaIngestionSpec{
		Type: "kafka",
		DataSchema: DataSchema{
			DataSource: "test",
			Parser: Parser{
				Type: "string",
				ParseSpec: ParseSpec{
					Format: "json",
					TimeStampSpec: TimestampSpec{
						Column: "timestamp",
						Format: "iso",
					},
					FlattenSpec: FlattenSpec{
						Fields: FieldList{
							Field{
								Type: "test",
								Name: "test",
								Expr: "test",
							},
						},
					},
					DimensionsSpec: DimensionsSpec{
						Dimensions: []string{"test"},
					},
				},
			},
			MetricsSpec: []Metric{
				{
					Name: "count",
					Type: "count",
				},
				{
					Name:      "value",
					FieldName: "value",
					Type:      "doubleMax",
				},
			},
			GranularitySpec: GranularitySpec{
				Type:               "uniform",
				SegmentGranularity: "HOUR",
				QueryGranularity:   "MINUTE",
			},
		},
		IOConfig: IOConfig{
			Topic: "test",
			ConsumerProperties: ConsumerProperties{
				BootstrapServers: "test",
			},
			TaskDuration:      "PT10M",
			UseEarliestOffset: true,
		},
	}
}

func TestKafkaIngestionSpec(t *testing.T) {
	var testData = []struct {
		name       string
		dataSource string
		topic      string
		brokers    string
		labels     LabelSet
		expected   *KafkaIngestionSpec
	}{
		{
			name:       "empty labels",
			dataSource: "test",
			topic:      "test",
			brokers:    "test",
			labels:     LabelSet{},
			expected: func() *KafkaIngestionSpec {
				out := defaultKafkaIngestionSpec()
				out.DataSchema.Parser.ParseSpec.FlattenSpec.Fields = FieldList{}
				out.DataSchema.Parser.ParseSpec.DimensionsSpec.Dimensions = []string{}
				return out
			}(),
		},
		{
			name:       "single label",
			dataSource: "test",
			topic:      "test",
			brokers:    "test",
			labels:     LabelSet{"foo"},
			expected: func() *KafkaIngestionSpec {
				out := defaultKafkaIngestionSpec()
				out.DataSchema.Parser.ParseSpec.FlattenSpec.Fields = FieldList{
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
				}
				out.DataSchema.Parser.ParseSpec.DimensionsSpec.Dimensions = []string{"name", "foo"}
				return out
			}(),
		},
		{
			name:       "multiple labels",
			dataSource: "test",
			topic:      "test",
			brokers:    "test",
			labels:     LabelSet{"foo", "bar"},
			expected: func() *KafkaIngestionSpec {
				out := defaultKafkaIngestionSpec()
				out.DataSchema.Parser.ParseSpec.FlattenSpec.Fields = FieldList{
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
				}
				out.DataSchema.Parser.ParseSpec.DimensionsSpec.Dimensions = []string{"name", "foo", "bar"}
				return out
			}(),
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			actual := NewKafkaIngestionSpec(
				test.dataSource,
				test.topic,
				test.brokers,
				test.labels,
			)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestKafkaIngestionSpec_MarshalJSON(t *testing.T) {
	spec := NewKafkaIngestionSpec(
		"test",
		"test",
		"test",
		LabelSet{"instance", "job"},
	)
	actual, err := json.MarshalIndent(spec, "", "    ")
	if err != nil {
		t.Fatalf("unexpected error while marshalling: %v", err)
	}
	expected := []byte(jsonComplete)
	assert.Equal(t, expected, actual, fmt.Sprintf("expected: %s\nactual: %s", string(expected), string(actual)))

	var checkSpec *KafkaIngestionSpec
	err = json.Unmarshal(actual, &checkSpec)
	if err != nil {
		t.Fatalf("unexpected error while unmarshalling: %v", err)
	}
	assert.Equal(t, spec, checkSpec)
}
