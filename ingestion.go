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

// KafkaIngestionSpec is the root-level type defining an ingestion spec used
// by Apache Druid.
type KafkaIngestionSpec struct {
	Type       string     `json:"type"`
	DataSchema DataSchema `json:"dataSchema"`
	IOConfig   IOConfig   `json:"ioConfig"`
}

// DataSchema represents the Druid dataSchema spec.
// Right now only the legacy spec is supported.
type DataSchema struct {
	DataSource      string          `json:"dataSource"`
	Parser          Parser          `json:"parser"`
	MetricsSpec     []Metric        `json:"metricsSpec"`
	GranularitySpec GranularitySpec `json:"granularitySpec"`
}

// Parser is responsible for configuring a wide variety of items related to
// parsing input records.
type Parser struct {
	Type      string    `json:"type"`
	ParseSpec ParseSpec `json:"parseSpec"`
}

// ParseSpec represents the parseSpec object under Parser.
type ParseSpec struct {
	Format         string         `json:"format"`
	TimeStampSpec  TimestampSpec  `json:"timestampSpec"`
	FlattenSpec    FlattenSpec    `json:"flattenSpec"`
	DimensionsSpec DimensionsSpec `json:"dimensionsSpec"`
}

// TimestampSpec is responsible for configuring the primary timestamp.
type TimestampSpec struct {
	Column string `json:"column"`
	Format string `json:"format"`
}

// FlattenSpec responsible for bridging the gap between potentially nested input
// data (such as JSON, Avro, etc) and Druid's flat data model.
type FlattenSpec struct {
	Fields FieldList `json:"fields"`
}

// DimensionsSpec is responsible for configuring Druid's dimensions. They're a
// set of columns in Druid's data model that can be used for grouping, filtering
// or applying aggregations.
type DimensionsSpec struct {
	Dimensions LabelSet `json:"dimensions"`
}

// FieldList is a list of Fields.
type FieldList []Field

// Field defines a piece of data.
type Field struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Expr string `json:"expr"`
}

// Metric is a Druid aggregator that is applied at ingestion time.
type Metric struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	FieldName string `json:"fieldName,omitempty"`
}

// GranularitySpec allows for configuring operations such as data segment
// partitioning, truncating timestamps, time chunk segmentation or roll-up.
type GranularitySpec struct {
	Type               string `json:"type"`
	SegmentGranularity string `json:"segmentGranularity"`
	QueryGranularity   string `json:"queryGranularity"`
}

// IOConfig influences how data is read into Druid from a source system. Right
// now only Kafka is supported.
type IOConfig struct {
	Topic              string                  `json:"topic"`
	ConsumerProperties KafkaConsumerProperties `json:"consumerProperties"`
	TaskDuration       string                  `json:"taskDuration"`
	UseEarliestOffset  bool                    `json:"useEarliestOffset"`
}

// KafkaConsumerProperties is a set of properties that is passed to the Kafka
// consumer.
type KafkaConsumerProperties struct {
	BootstrapServers      string            `json:"bootstrap.servers"`
	SecurityProtocol      *string           `json:"security.protocol,omitempty"`
	SSLTruststoreType     *string           `json:"ssl.truststore.type,omitempty"`
	SSLEnabledProtocols   *string           `json:"ssl.enabled.protocols,omitempty"`
	SSLTruststoreLocation *string           `json:"ssl.truststore.location,omitempty"`
	SSLTruststorePassword *PasswordProvider `json:"ssl.truststore.password,omitempty"`
	SSLKeystoreLocation   *string           `json:"ssl.keystore.location,omitempty"`
	SSLKeystorePassword   *PasswordProvider `json:"ssl.keystore.password,omitempty"`
}

// PasswordProvider allows Druid to configure secrets via environment variables.
type PasswordProvider struct {
	Type     string `json:"type"`
	Variable string `json:"variable"`
}

// defaultKafkaIngestionSpec returns a default KafkaIngestionSpec
func defaultKafkaIngestionSpec() *KafkaIngestionSpec {
	spec := &KafkaIngestionSpec{
		Type: "kafka",
		DataSchema: DataSchema{
			DataSource: "prometheus",
			Parser: Parser{
				Type: "string",
				ParseSpec: ParseSpec{
					Format: "json",
					TimeStampSpec: TimestampSpec{
						Column: "timestamp",
						Format: "iso",
					},
					FlattenSpec: FlattenSpec{
						Fields: FieldList{},
					},
					DimensionsSpec: DimensionsSpec{
						Dimensions: []string{},
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
					Type:      "doubleMax",
					FieldName: "value",
				},
			},
			GranularitySpec: GranularitySpec{
				Type:               "uniform",
				SegmentGranularity: "HOUR",
				QueryGranularity:   "MINUTE",
			},
		},
		IOConfig: IOConfig{
			Topic: "prometheus",
			ConsumerProperties: KafkaConsumerProperties{
				BootstrapServers: "kafka01:9090,kafka02:9090,kafka03:9090",
			},
			TaskDuration:      "PT10M",
			UseEarliestOffset: true,
		},
	}
	return spec
}

// NewKafkaIngestionSpec returns a default KafkaIngestionSpec and applies any
// options passed to it.
func NewKafkaIngestionSpec(options ...KafkaIngestionSpecOptions) *KafkaIngestionSpec {
	spec := defaultKafkaIngestionSpec()
	for _, fn := range options {
		fn(spec)
	}
	return spec
}
