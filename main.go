package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/jamiealquiza/cloudwatch-graphite/vendor/github.com/awslabs/aws-sdk-go/aws"
	"github.com/jamiealquiza/cloudwatch-graphite/vendor/github.com/awslabs/aws-sdk-go/service/cloudwatch"
)

var (
	flagRegion         string
	flagNamespace      string
	flagPeriod         int64
	flagFetchPrevious  int
	flagDimensionName  string
	flagDimensionValue string
	flagMetrics        string
	flagList           bool
	flagDump           bool

	dimensionNameRe  *regexp.Regexp
	dimensionValueRe *regexp.Regexp
	metricsRe        *regexp.Regexp

	// ...Because instantiating pointer literals.
	statisticsPointers = []*string{}
	statistics         = []string{"Average"}
)

func init() {
	flag.StringVar(&flagRegion, "region", "us-east-1", "AWS region")
	flag.StringVar(&flagNamespace, "namespace", "", "Namespace")
	flag.Int64Var(&flagPeriod, "period", 1, "Period (multiples of 60s)")
	flag.IntVar(&flagFetchPrevious, "fetch-previous", 60, "Negative time in minutes from now to fetch metrics")
	flag.StringVar(&flagDimensionName, "dimension-name", "", "Dimension Name regex")
	flag.StringVar(&flagDimensionValue, "dimension-value", "", "Dimension Value regex")
	flag.StringVar(&flagMetrics, "metrics", "", "Metrics name regex")
	flag.BoolVar(&flagList, "list", false, "Print metrics available matching filter, but don't fetch")
	flag.BoolVar(&flagDump, "dump", true, "Dump raw metrics data received")
	flag.Parse()

	if flagDimensionName != "" {
		dimensionNameRe = regexp.MustCompile(flagDimensionName)
	}
	if flagDimensionValue != "" {
		dimensionValueRe = regexp.MustCompile(flagDimensionValue)
	}

	if flagMetrics != "" {
		metricsRe = regexp.MustCompile(flagMetrics)
	} else {
		metricsRe = regexp.MustCompile(".*")
	}

	// flag.StringVar(&statistics, ) Needs to take list, turn into slice.
	for _, stat := range statistics {
		statisticsPointers = append(statisticsPointers, &stat)
	}

}

func getAvailableMetrics(cw *cloudwatch.CloudWatch, lmr *cloudwatch.ListMetricsInput) ([]*cloudwatch.Metric, error) {
	metricsToRequest := make([]*cloudwatch.Metric, 0)

contList:
	resp, err := cw.ListMetrics(lmr)
	if err != nil {
		return nil, err
	}

	lmr.NextToken = resp.NextToken

	for _, metric := range resp.Metrics {

		// Check agains metrics filter.
		if !metricFilter(metric.MetricName) {
			continue
		}

		// Although, we do an allocation here that may
		// get axed after the dimension filter.
		m := &cloudwatch.Metric{
			MetricName: metric.MetricName,
			Namespace:  metric.Namespace,
		}

		// Check if dimensions is empty, then filter.
		if len(metric.Dimensions) > 0 {
			for _, d := range metric.Dimensions {
				if dimensionFilter(d) {
					m.Dimensions = append(m.Dimensions, d)
				}
			}

			if len(m.Dimensions) < 1 {
				continue
			}

			metricsToRequest = append(metricsToRequest, m)
		}
	}

	// If we got a NextToken, fetch the next
	// sequence.
	if lmr.NextToken != nil {
		goto contList
	}

	return metricsToRequest, nil
}

func fetchMetrics(cw *cloudwatch.CloudWatch, am []*cloudwatch.Metric) ([][]byte, error) {
	metricsFetched := make([][]byte, 0)

	endTs := time.Now()
	startTs := endTs.Add(-time.Duration(flagFetchPrevious) * time.Minute)
	period := flagPeriod * 60

	for _, m := range am {

		// Init metric request body.
		getMetricsRequest := &cloudwatch.GetMetricStatisticsInput{
			Namespace:  m.Namespace,
			Dimensions: m.Dimensions,
			MetricName: m.MetricName,
			Period:     &period,
			Statistics: statisticsPointers,
			EndTime:    &endTs,
			StartTime:  &startTs,
		}

		resp, err := cw.GetMetricStatistics(getMetricsRequest)
		if err != nil {
			return nil, err
		}

		j, err := json.Marshal(resp)
		if err != nil {
			return nil, err
		}

		metricsFetched = append(metricsFetched, j)

	}

	return metricsFetched, nil
}

func metricFilter(s *string) bool {
	// No filter set, short circuit.
	if flagMetrics == "" {
		return true
	}

	if metricsRe.MatchString(*s) {
		return true
	}

	return false
}

func dimensionFilter(d *cloudwatch.Dimension) bool {
	// No filters set, short circuit.
	if dimensionNameRe == nil && dimensionValueRe == nil {
		return true
	}

	// To reduce the logic here, if only one dimension filter
	// is set, set the other to ".*". We do this because we want
	// the dimension Name or Value values to match all filter conditions,
	// assuming that any single filter left unset means ".*".
	if dimensionNameRe == nil {
		dimensionNameRe = regexp.MustCompile(".*")
	}

	if dimensionValueRe == nil {
		dimensionValueRe = regexp.MustCompile(".*")
	}

	if dimensionNameRe.MatchString(*d.Name) && dimensionValueRe.MatchString(*d.Value) {
		return true
	}

	return false
}

// stringMetric is used to format the available metrics output
// when the --list flag is set.
func stringMetric(m *cloudwatch.Metric) string {
	var dimensions string

	for _, d := range m.Dimensions {
		dimensions += fmt.Sprintf("%s=%s", *d.Name, *d.Value)
	}

	return fmt.Sprintf("Dimensions: %s, MetricName: %s", dimensions, *m.MetricName)
}

func main() {
	cw := cloudwatch.New(&aws.Config{Region: flagRegion})

	listMetricsRequest := &cloudwatch.ListMetricsInput{Namespace: &flagNamespace}
	availableMetrics, _ := getAvailableMetrics(cw, listMetricsRequest)

	if flagList {
		for _, m := range availableMetrics {
			fmt.Println(stringMetric(m))
		}
		os.Exit(0)
	}

	metrics, err := fetchMetrics(cw, availableMetrics)
	if err != nil {
		fmt.Println(err)
	}

	if flagDump {
		for _, m := range metrics {
			fmt.Println(string(m))
		}
		os.Exit(0)
	}

}
