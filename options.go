package prometheus_kafka_adapter_druid_ingestion

type KafkaIngestionSpecOptions func(*KafkaIngestionSpec)

func stringPointer(s string) *string {
	return &s
}

func ApplySSLConfig() KafkaIngestionSpecOptions {
	return func(spec *KafkaIngestionSpec) {
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
}

func SetDataSource(ds string) KafkaIngestionSpecOptions {
	return func(spec *KafkaIngestionSpec) {
		spec.DataSchema.DataSource = ds
	}
}

func SetTopic(topic string) KafkaIngestionSpecOptions {
	return func(spec *KafkaIngestionSpec) {
		spec.IOConfig.Topic = topic
	}
}

func SetBrokers(brokers string) KafkaIngestionSpecOptions {
	return func(spec *KafkaIngestionSpec) {
		spec.IOConfig.ConsumerProperties.BootstrapServers = brokers
	}
}

func SetLabels(labels LabelSet) KafkaIngestionSpecOptions {
	return func(spec *KafkaIngestionSpec) {
		spec.DataSchema.Parser.ParseSpec.FlattenSpec.Fields = labels.ToFieldList()
		spec.DataSchema.Parser.ParseSpec.DimensionsSpec.Dimensions = labels.ToDimensions()
	}
}
