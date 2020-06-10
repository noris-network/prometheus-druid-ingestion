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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	jsonBasic = `{
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
	jsonSSL = `{
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
            "bootstrap.servers": "test",
            "security.protocol": "SSL",
            "ssl.truststore.type": "PKCS12",
            "ssl.enabled.protocols": "TLSv1.2",
            "ssl.truststore.location": "/var/private/ssl/truststore.p12",
            "ssl.truststore.password": {
                "type": "environment",
                "variable": "DRUID_TRUSTSTORE_PASSWORD"
            },
            "ssl.keystore.location": "/var/private/ssl/keystore.p12",
            "ssl.keystore.password": {
                "type": "environment",
                "variable": "DRUID_KEYSTORE_PASSWORD"
            }
        },
        "taskDuration": "PT10M",
        "useEarliestOffset": true
    }
}`
)

func TestKafkaIngestionSpec(t *testing.T) {
	var testData = []struct {
		name     string
		options  []KafkaIngestionSpecOptions
		expected *KafkaIngestionSpec
	}{
		{
			name: "empty labels",
			options: []KafkaIngestionSpecOptions{
				SetDataSource("test"),
				SetTopic("test"),
				SetBrokers("test"),
				SetLabels(LabelSet{}),
			},
			expected: func() *KafkaIngestionSpec {
				out := defaultKafkaIngestionSpec()
				out.DataSchema.Parser.ParseSpec.FlattenSpec.Fields = FieldList{}
				out.DataSchema.Parser.ParseSpec.DimensionsSpec.Dimensions = []string{}
				out.DataSchema.DataSource = "test"
				out.IOConfig.Topic = "test"
				out.IOConfig.ConsumerProperties.BootstrapServers = "test"
				return out
			}(),
		},
		{
			name: "empty labels, ssl options",
			options: []KafkaIngestionSpecOptions{
				ApplySSLConfig(),
				SetDataSource("test"),
				SetTopic("test"),
				SetBrokers("test"),
				SetLabels(LabelSet{}),
			},
			expected: func() *KafkaIngestionSpec {
				out := defaultKafkaIngestionSpec()
				out.DataSchema.Parser.ParseSpec.FlattenSpec.Fields = FieldList{}
				out.DataSchema.Parser.ParseSpec.DimensionsSpec.Dimensions = []string{}
				out.DataSchema.DataSource = "test"
				out.IOConfig.Topic = "test"
				out.IOConfig.ConsumerProperties.BootstrapServers = "test"
				out.IOConfig.ConsumerProperties.SecurityProtocol = stringPointer("SSL")
				out.IOConfig.ConsumerProperties.SSLTruststoreType = stringPointer("PKCS12")
				out.IOConfig.ConsumerProperties.SSLEnabledProtocols = stringPointer("TLSv1.2")
				out.IOConfig.ConsumerProperties.SSLTruststoreLocation = stringPointer("/var/private/ssl/truststore.p12")
				out.IOConfig.ConsumerProperties.SSLTruststorePassword = &PasswordProvider{
					Type:     "environment",
					Variable: "DRUID_TRUSTSTORE_PASSWORD",
				}
				out.IOConfig.ConsumerProperties.SSLKeystoreLocation = stringPointer("/var/private/ssl/keystore.p12")
				out.IOConfig.ConsumerProperties.SSLKeystorePassword = &PasswordProvider{
					Type:     "environment",
					Variable: "DRUID_KEYSTORE_PASSWORD",
				}
				return out
			}(),
		},
		{
			name: "single label",
			options: []KafkaIngestionSpecOptions{
				SetDataSource("test"),
				SetTopic("test"),
				SetBrokers("test"),
				SetLabels(LabelSet{"foo"}),
			},
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
				out.DataSchema.DataSource = "test"
				out.IOConfig.Topic = "test"
				out.IOConfig.ConsumerProperties.BootstrapServers = "test"
				return out
			}(),
		},
		{
			name: "multiple labels, ssl config",
			options: []KafkaIngestionSpecOptions{
				ApplySSLConfig(),
				SetDataSource("test"),
				SetTopic("test"),
				SetBrokers("test"),
				SetLabels(LabelSet{"foo", "bar"}),
			},
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
				out.DataSchema.DataSource = "test"
				out.IOConfig.Topic = "test"
				out.IOConfig.ConsumerProperties.BootstrapServers = "test"
				out.IOConfig.ConsumerProperties.SecurityProtocol = stringPointer("SSL")
				out.IOConfig.ConsumerProperties.SSLTruststoreType = stringPointer("PKCS12")
				out.IOConfig.ConsumerProperties.SSLEnabledProtocols = stringPointer("TLSv1.2")
				out.IOConfig.ConsumerProperties.SSLTruststoreLocation = stringPointer("/var/private/ssl/truststore.p12")
				out.IOConfig.ConsumerProperties.SSLTruststorePassword = &PasswordProvider{
					Type:     "environment",
					Variable: "DRUID_TRUSTSTORE_PASSWORD",
				}
				out.IOConfig.ConsumerProperties.SSLKeystoreLocation = stringPointer("/var/private/ssl/keystore.p12")
				out.IOConfig.ConsumerProperties.SSLKeystorePassword = &PasswordProvider{
					Type:     "environment",
					Variable: "DRUID_KEYSTORE_PASSWORD",
				}
				return out
			}(),
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			actual := NewKafkaIngestionSpec(
				test.options...,
			)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestKafkaIngestionSpec_MarshalJSON(t *testing.T) {
	t.Run("jsonBasic", func(t *testing.T) {
		spec := NewKafkaIngestionSpec(
			SetDataSource("test"),
			SetTopic("test"),
			SetBrokers("test"),
			SetLabels(LabelSet{"instance", "job"}),
		)
		actual, err := json.MarshalIndent(spec, "", "    ")
		if err != nil {
			t.Fatalf("unexpected error while marshalling: %v", err)
		}
		expected := []byte(jsonBasic)
		assert.Equal(t, string(expected), string(actual), fmt.Sprintf("expected: %s\nactual: %s", string(expected), string(actual)))

		var checkSpec *KafkaIngestionSpec
		err = json.Unmarshal(actual, &checkSpec)
		if err != nil {
			t.Fatalf("unexpected error while unmarshalling: %v", err)
		}
		assert.Equal(t, spec, checkSpec)
	})

	t.Run("jsonSSL", func(t *testing.T) {
		spec := NewKafkaIngestionSpec(
			ApplySSLConfig(),
			SetDataSource("test"),
			SetTopic("test"),
			SetBrokers("test"),
			SetLabels(LabelSet{"instance", "job"}),
		)
		actual, err := json.MarshalIndent(spec, "", "    ")
		if err != nil {
			t.Fatalf("unexpected error while marshalling: %v", err)
		}
		expected := []byte(jsonSSL)
		assert.Equal(t, string(expected), string(actual), fmt.Sprintf("expected: %s\nactual: %s", string(expected), string(actual)))

		var checkSpec *KafkaIngestionSpec
		err = json.Unmarshal(actual, &checkSpec)
		if err != nil {
			t.Fatalf("unexpected error while unmarshalling: %v", err)
		}
		assert.Equal(t, spec, checkSpec)
	})
}

var result *KafkaIngestionSpec

func BenchmarkNewKafkaIngestionSpec(b *testing.B) {
	var spec *KafkaIngestionSpec
	for i := 0; i < b.N; i++ {
		spec = NewKafkaIngestionSpec(
			SetDataSource("raw.prometheus"),
			SetBrokers("kafka01:9092,kafka02:9092"),
			SetLabels(LabelSet{"job", "instance", "source", "namespace"}),
			ApplySSLConfig(),
		)
	}
	result = spec
}

var resultJSON []byte

func BenchmarkNewKafkaIngestionSpecMarshalJSON(b *testing.B) {
	var (
		spec   *KafkaIngestionSpec
		actual []byte
		err    error
	)

	for i := 0; i < b.N; i++ {
		spec = NewKafkaIngestionSpec(
			SetDataSource("raw.prometheus"),
			SetBrokers("kafka01:9092,kafka02:9092"),
			SetLabels(LabelSet{"job", "instance", "source", "namespace"}),
			ApplySSLConfig(),
		)
		actual, err = json.MarshalIndent(spec, "", "    ")
		if err != nil {
			b.Fatal(err)
		}
	}
	result = spec
	resultJSON = actual
}
