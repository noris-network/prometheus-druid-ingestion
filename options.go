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

// KafkaIngestionSpecOptions allows for configuring a KafkaIngestionSpec.
type KafkaIngestionSpecOptions func(*KafkaIngestionSpec)

func stringPointer(s string) *string {
	return &s
}

// ApplySSLConfig adds an opinionated SSL config that is used for communicating
// with Kafka securely.
func ApplySSLConfig() KafkaIngestionSpecOptions {
	return func(spec *KafkaIngestionSpec) {
		spec.IOConfig.ConsumerProperties.SecurityProtocol = stringPointer("SSL")
		spec.IOConfig.ConsumerProperties.SSLTruststoreType = stringPointer("PKCS12")
		spec.IOConfig.ConsumerProperties.SSLEnabledProtocols = stringPointer("TLSv1.2")
		spec.IOConfig.ConsumerProperties.SSLTruststoreLocation = stringPointer("/var/private/ssl/truststore.p12")
		spec.IOConfig.ConsumerProperties.SSLTruststorePassword = &PasswordProvider{
			Type:     "environment",
			Variable: "DRUID_TRUSTSTORE_PASSWORD",
		}
		spec.IOConfig.ConsumerProperties.SSLKeystoreLocation = stringPointer("/var/private/ssl/keystore.p12")
		spec.IOConfig.ConsumerProperties.SSLKeystorePassword = &PasswordProvider{
			Type:     "environment",
			Variable: "DRUID_KEYSTORE_PASSWORD",
		}
	}
}

// SetDataSource sets the name of the dataSource used in Druid.
func SetDataSource(ds string) KafkaIngestionSpecOptions {
	return func(spec *KafkaIngestionSpec) {
		spec.DataSchema.DataSource = ds
	}
}

// SetTopic sets the Kafka topic to consume data from.
func SetTopic(topic string) KafkaIngestionSpecOptions {
	return func(spec *KafkaIngestionSpec) {
		spec.IOConfig.Topic = topic
	}
}

// SetBrokers sets the addresses of Kafka brokers. E.g. 'kafka01:9092,
// kafka02:9092,kafka03:9092'.
func SetBrokers(brokers string) KafkaIngestionSpecOptions {
	return func(spec *KafkaIngestionSpec) {
		spec.IOConfig.ConsumerProperties.BootstrapServers = brokers
	}
}

// SetLabels uses a LabelSet to configure the ingestion spec with.
// This sets the FieldList under FlattenSpec, as well as Dimensions.
func SetLabels(labels LabelSet) KafkaIngestionSpecOptions {
	return func(spec *KafkaIngestionSpec) {
		spec.DataSchema.Parser.ParseSpec.FlattenSpec.Fields = labels.ToFieldList()
		spec.DataSchema.Parser.ParseSpec.DimensionsSpec.Dimensions = labels.ToDimensions()
	}
}
