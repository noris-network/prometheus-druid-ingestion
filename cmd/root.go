package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	ingestion "github.com/noris-network/prometheus-kafka-adapter-druid-ingestion"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	address         = "http://prometheus:9090"
	query           = `{__name__=~"job:.+"}`
	tlsSkipVerify   = false
	toStdout        = true
	outputFile      = ""
	druidDataSource = "prometheus"
	kafkaTopic      = "prometheus"
	kafkaBrokers    = "kafka01:9092,kafka02:9092,kafka03:9092"
	ingestSSL       = true
	rootCmd         = &cobra.Command{
		Use:   "generate-ingestion",
		Short: "Generate an Druid.io opinionated ingestion spec from a Prometheus query result",
		Run:   run,
	}
)

func init() {
	f := rootCmd.Flags()
	f.StringVarP(&address, "address", "a", address, "The address of the Prometheus server to send the query to")
	f.StringVarP(&query, "query", "q", query, "The query to send to the Prometheus server")
	f.BoolVar(&tlsSkipVerify, "tls-skip-verify", tlsSkipVerify, "Skip TLS certificate verification")
	f.BoolVarP(&toStdout, "toStdout", "o", toStdout, "Prints the JSON ingestion spec to STDOUT")
	f.StringVarP(&outputFile, "file", "f", outputFile, "The file to save the ingestion spec to")
	f.StringVarP(&druidDataSource, "druid-data-source", "d", druidDataSource, "The druid data source")
	f.StringVarP(&kafkaTopic, "kafka-topic", "t", kafkaTopic, "The Kafka topic for druid to ingest data from")
	f.StringVarP(&kafkaBrokers, "kafka-brokers", "b", kafkaBrokers, "The Kafka brokers for druid to ingest data from")
	f.BoolVar(&ingestSSL, "ingest-via-ssl", ingestSSL, "Enables data ingestion from Kafka to Druid via SSL")
}

func Run() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	rt := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: tlsSkipVerify,
		},
	}
	client, err := api.NewClient(api.Config{
		Address:      address,
		RoundTripper: rt,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, fmt.Sprintf(`%s`, query), time.Now())
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	l, err := ingestion.ExtractUniqueLabels(result)
	if err != nil {
		fmt.Printf("Error extracting labels: %v\n", err)
		os.Exit(1)
	}

	var opts []ingestion.KafkaIngestionSpecOptions
	if ingestSSL {
		opts = append(opts, ingestion.ApplySSLConfig)
	}

	spec := ingestion.NewKafkaIngestionSpec(druidDataSource, kafkaTopic, kafkaBrokers, l, opts...)
	jsonSpec, err := json.MarshalIndent(spec, "", "    ")
	if err != nil {
		fmt.Printf("Error marshalling ingestion spec: %v\n", err)
		os.Exit(1)
	}

	if toStdout {
		fmt.Println(string(jsonSpec))
	}
	if outputFile != "" {
		if err = ioutil.WriteFile(outputFile, jsonSpec, os.FileMode(0644)); err != nil {
			fmt.Printf("Error writing %q: %v", outputFile, err)
			os.Exit(1)
		}
	}
}
