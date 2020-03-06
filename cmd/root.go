package main

import "github.com/spf13/cobra"

// Default variables
var (
	address       = "http://prometheus:9090"
	query         = `{__name__=~"job:.+"}`
	tlsSkipVerify = false
	tlsCertPath   = ""
	rootCmd       = &cobra.Command{
		Use:   "spec",
		Short: "Generate a Druid.io IngestionSpec from a Prometheus query result.",
	}
)

func init() {
	f := rootCmd.Flags()
	f.StringVarP(&address, "address", "a", address, "The address of the Prometheus server to send the query to")
	f.StringVarP(&query, "query", "q", query, "The query to send to the Prometheus server")
	f.BoolVar(&tlsSkipVerify, "tls-skip-verify", tlsSkipVerify, "Skip TLS certificate verification")
	f.StringVar(&tlsCertPath, "tls-cert-path", tlsCertPath, "Path to the TLS certificate")
}
