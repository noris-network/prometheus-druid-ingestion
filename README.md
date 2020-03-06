# prometheus-kafka-druid-ingestion

At [noris network](https://noris.de) we're sending Prometheus data via `[prometheus-kafka-adapter][pka] to 
Kafka and then use [Apache Druid][druid]'s [Kafka Ingestion][kafka_ingestion]
for storing those metrics.

This repo creates an [Apache Druid][druid] [ingestion spec][ingestion_spec] from [prometheus-kafka-adapter][pka] data.

## Background

Prometheus data can be sent via [remote_write](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write) to 
[prometheus-kafka-adapter][pka], which in turn sends it to Kafka. The prometheus-kafka-adapter message will have the following structure:

```json
{
  "timestamp": "1970-01-01T00:00:00Z",
  "value": "9876543210",
  "name": "up",

  "labels": {
    "__name__": "up",
    "label1": "value1",
    "label2": "value2"
  }
}
```

This data can be send to [Apache Druid][druid] for long term storage, using Druid's [Kafka Ingestion][kafka_ingestion] feature. For the
data to be consumed and saved, we need to create an [ingestion spec][ingestion_spec], including the final schema in the database
and the Kafka brokers to consume the metrics from.

## Usage

```text
$ ./bin/generate-ingestion -h

Usage:
  generate-ingestion [flags]

Flags:
  -a, --address string             The address of the Prometheus server to send the query to (default "http://prometheus:9090")
  -d, --druid-data-source string   The druid data source (default "prometheus")
  -f, --file string                The file to save the ingestion spec to
  -h, --help                       help for generate-ingestion
      --ingest-via-ssl             Enables data ingestion from Kafka to Druid via SSL (default true)
  -b, --kafka-brokers string       The Kafka brokers for druid to ingest data from (default "kafka01:9092,kafka02:9092,kafka03:9092")
  -t, --kafka-topic string         The Kafka topic for druid to ingest data from (default "prometheus")
  -q, --query string               The query to send to the Prometheus server (default "{__name__=~\"job:.+\"}")
      --tls-skip-verify            Skip TLS certificate verification
  -o, --toStdout                   Prints the JSON ingestion spec to STDOUT (default true)
```

Executing the file sends the query specified with the `-q` / `--query` flag to a Prometheus server
running at `-a` / `--address`. The script will then extract the unique labels of all time series returned
and build an opinionated ingestion spec from those labels.

> If all recording rules with the prefix `job:` are sent to [prometheus-kafka-adapter][pka] via `remote_write`,
> the PromQL query `{__name__=~"job:.+"}` would retrieve those series.

By default Druid will consume messages from Kafka via SSL, meaning the following block will
be populated in the spec:

```text
            # ...
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
```

This behaviour can be disabled with the `--ingest-via-ssl=false` flag.

By default the ingestion spec is displayed to `stdout`, but can be saved with the `-f` flag:

```text
$ generate-ingestion -f ingestion.json
```
```json
{
    "type": "kafka",
    "dataSchema": {
        "dataSource": "prometheus",
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
                            "name": "clustername",
                            "expr": "$.labels.clustername"
                        },
                        {
                            "type": "path",
                            "name": "source",
                            "expr": "$.labels.source"
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
                        "clustername",
                        "source"
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
        "topic": "prometheus",
        "consumerProperties": {
            "bootstrap.servers": "kafka01:9092,kafka02:9092,kafka03:9092"
        },
        "taskDuration": "PT10M",
        "useEarliestOffset": true
    }
}
```

[pka]: https://github.com/Telefonica/prometheus-kafka-adapter
[druid]: https://druid.apache.org
[ingestion_spec]: https://druid.apache.org/docs/latest/ingestion/index.html
[kafka_ingestion]: https://druid.apache.org/docs/latest/development/extensions-core/kafka-ingestion.html