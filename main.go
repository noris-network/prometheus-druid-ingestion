package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"net"
	"net/http"
	"os"
	"time"
)

type LabelSet []string

func extractUniqueLabels(result model.Value) (LabelSet, error) {
	vec, ok := result.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("query result is not a Vector")
	}

	labelSeen := make(map[string]bool)
	var labels LabelSet
	for _, m := range vec {
		for k := range m.Metric {
			key := string(k)
			if _, ok := labelSeen[key]; !ok {
				fmt.Printf("Adding %s", key)
				labels = append(labels, key)
				labelSeen[key] = true
			}
		}
	}
	return labels, nil
}

func main() {
	rt := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client, err := api.NewClient(api.Config{
		Address:      "https://prometheus.capman-test.noris.de",
		RoundTripper: rt,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, `{__name__=~"nude:.+"}`, time.Now())
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	labels, err := extractUniqueLabels(result)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Unique labels: %v\n", labels)
}
