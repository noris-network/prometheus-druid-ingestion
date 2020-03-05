package main

type KafkaIngestionSpec struct {
	Type       string     `json:"type"`
	DataSchema DataSchema `json:"dataSchema"`
	IOConfig   IOConfig   `json:"ioConfig"`
}

type DataSchema struct {
	DataSource      string          `json:"dataSource"`
	Parser          Parser          `json:"parser"`
	MetricsSpec     []Metric        `json:"metricsSpec"`
	GranularitySpec GranularitySpec `json:"granularitySpec"`
}

type Parser struct {
	Type      string    `json:"type"`
	ParseSpec ParseSpec `json:"parseSpec"`
}

type ParseSpec struct {
	Format         string         `json:"format"`
	TimeStampSpec  TimestampSpec  `json:"timestampSpec"`
	FlattenSpec    FlattenSpec    `json:"flattenSpec"`
	DimensionsSpec DimensionsSpec `json:"dimensionsSpec"`
}

type TimestampSpec struct {
	Column string `json:"column"`
	Format string `json:"format"`
}

type FlattenSpec struct {
	Fields FieldList `json:"fields"`
}

type DimensionsSpec struct {
	Dimensions LabelSet `json:"dimensions"`
}

type FieldList []Field

type Field struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Expr string `json:"expr"`
}

type Metric struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	FieldName string `json:"fieldName,omitempty"`
}

type GranularitySpec struct {
	Type               string `json:"type"`
	SegmentGranularity string `json:"segmentGranularity"`
	QueryGranularity   string `json:"queryGranularity"`
}

type IOConfig struct {
	Topic              string             `json:"topic"`
	ConsumerProperties ConsumerProperties `json:"consumerProperties"`
	TaskDuration       string             `json:"taskDuration"`
	UseEarliestOffset  bool               `json:"useEarliestOffset"`
}

type ConsumerProperties struct {
	BootstrapServers string `json:"bootstrap.servers"`
}

func NewKafkaIngestionSpec(dataSource, topic, brokers string, labels LabelSet) *KafkaIngestionSpec {
	return &KafkaIngestionSpec{
		Type: "kafka",
		DataSchema: DataSchema{
			DataSource: dataSource,
			Parser: Parser{
				Type: "string",
				ParseSpec: ParseSpec{
					Format: "json",
					TimeStampSpec: TimestampSpec{
						Column: "timestamp",
						Format: "iso",
					},
					FlattenSpec: FlattenSpec{
						Fields: labels.ToFieldList(),
					},
					DimensionsSpec: DimensionsSpec{
						Dimensions: labels.ToDimensions(),
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
			Topic: topic,
			ConsumerProperties: ConsumerProperties{
				BootstrapServers: brokers,
			},
			TaskDuration:      "PT10M",
			UseEarliestOffset: true,
		},
	}
}
