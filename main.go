// Copyright 2022 Yoshi Yamaguchi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

var (
	md      string
	project string
	short   bool
)

func init() {
	flag.StringVar(&project, "project", "", "Google Cloud Project ID")
	flag.StringVar(&md, "md", "", "Metric Descriptor name")
	flag.BoolVar(&short, "short", true, "Enable short output")
}

func main() {
	flag.Parse()

	if md == "" {
		if err := listMetric(project); err != nil {
			fmt.Printf("error listing metric descriptors: %v", err)
		}
		return
	}

	if err := deleteMetric(project, md); err != nil {
		fmt.Printf("failed to delete metric %v: %v\n", md, err)
	}
}

func deleteMetric(project, name string) error {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	fqdn := fmt.Sprintf("projects/%s/metricDescriptors/%s", project, name)
	req := &monitoringpb.DeleteMetricDescriptorRequest{
		Name: fqdn,
	}

	if err := c.DeleteMetricDescriptor(ctx, req); err != nil {
		return fmt.Errorf("could not delete metric: %v", err)
	}
	fmt.Printf("Deleted metric: %q\n", name)
	return nil
}

func listMetric(project string) error {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	req := &monitoringpb.ListMetricDescriptorsRequest{
		Name: fmt.Sprintf("projects/%s", project),
	}

	iter := c.ListMetricDescriptors(ctx, req)
	for {
		resp, err := iter.Next()
		if err != nil {
			return err
		}
		if isCustomMetric(resp.Type) {
			printMetricDescriptor(resp)
		}
	}
}

func isCustomMetric(name string) bool {
	return strings.HasPrefix(name, "custom.googleapis.com/") || strings.HasPrefix(name, "workload.googleapis.com/")
}

func printMetricDescriptor(p *metric.MetricDescriptor) {
	typ := p.Type
	if short {
		fmt.Printf("%v\n", typ)
		return
	}

	kind := p.MetricKind
	value := p.ValueType
	unit := strings.TrimSpace(p.Unit)
	desc := strings.TrimSpace(p.Description)

	fmt.Printf(`Type: %v
    Kind: %v
    Value: %v
    Unit: %v
    Description: %v
`, typ, kind, value, unit, desc)
}
