package prometheus_kafka_adapter_druid_ingestion

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
	BootstrapServers      string                 `json:"bootstrap.servers"`
	SecurityProtocol      *string                `json:"security.protocol,omitempty"`
	SSLTruststoreType     *string                `json:"ssl.truststore.type,omitempty"`
	SSLEnabledProtocols   *string                `json:"ssl.enabled.protocols,omitempty"`
	SSLTruststoreLocation *string                `json:"ssl.truststore.location,omitempty"`
	SSLTruststorePassword *DruidPasswordProvider `json:"ssl.truststore.password,omitempty"`
	SSLKeystoreLocation   *string                `json:"ssl.keystore.location,omitempty"`
	SSLKeystorePassword   *DruidPasswordProvider `json:"ssl.keystore.password,omitempty"`
}

type DruidPasswordProvider struct {
	Type     string `json:"type"`
	Variable string `json:"variable"`
}

type KafkaIngestionSpecOptions func(*KafkaIngestionSpec)

func stringPointer(s string) *string {
	return &s
}

func ApplySSLConfig(spec *KafkaIngestionSpec) {
	spec.IOConfig.ConsumerProperties.SecurityProtocol = stringPointer("SSL")
	spec.IOConfig.ConsumerProperties.SSLTruststoreType = stringPointer("PKCS12")
	spec.IOConfig.ConsumerProperties.SSLEnabledProtocols = stringPointer("TLSv1.2")
	spec.IOConfig.ConsumerProperties.SSLTruststoreLocation = stringPointer("/var/private/ssl/truststore.p12")
	spec.IOConfig.ConsumerProperties.SSLTruststorePassword = &DruidPasswordProvider{
		Type:     "environment",
		Variable: "DRUID_TRUSTSTORE_PASSWORD",
	}
	spec.IOConfig.ConsumerProperties.SSLKeystoreLocation = stringPointer("/var/private/ssl/keystore.p12")
	spec.IOConfig.ConsumerProperties.SSLKeystorePassword = &DruidPasswordProvider{
		Type:     "environment",
		Variable: "DRUID_KEYSTORE_PASSWORD",
	}
}

func NewKafkaIngestionSpec(dataSource, topic, brokers string, labels LabelSet, options ...KafkaIngestionSpecOptions) *KafkaIngestionSpec {
	spec := defaultKafkaIngestionSpec()
	spec.DataSchema.DataSource = dataSource
	spec.IOConfig.Topic = topic
	spec.IOConfig.ConsumerProperties.BootstrapServers = brokers
	spec.DataSchema.Parser.ParseSpec.FlattenSpec.Fields = labels.ToFieldList()
	spec.DataSchema.Parser.ParseSpec.DimensionsSpec.Dimensions = labels.ToDimensions()
	for _, fn := range options {
		fn(spec)
	}
	return spec
}

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
			ConsumerProperties: ConsumerProperties{
				BootstrapServers: "kafka01:9090,kafka02:9090,kafka03:9090",
			},
			TaskDuration:      "PT10M",
			UseEarliestOffset: true,
		},
	}
	return spec
}
