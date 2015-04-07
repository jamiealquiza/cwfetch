# cwfetch

WIP CloudWatch metrics fetching tool. Output to Graphite (and others) pending. 

cwfetch allows you to specify a CloudWatch metrics namespace and selectively fetch metrics data using regex filters on both dimension fields and metric names. Using the `-list` directive will only list CloudWatch metrics that will be fetched according to any filters that may be set, without actually fetching the metrics data. This allows you to iteratively define and test what metrics data will be fetched.

### Install

Assuming Go is installed:
 - `go get github.com/jamiealquiza/cwfetch`
 - `go install github.com/jamiealquiza/cwfetch`
 - Binary will be found at $GOPATH/bin/cwfetch

### Usage

Set AWS_ACCESS_KEY and AWS_SECRET_KEY env vars.

<pre>
% cwfetch -h
Usage of cwfetch:
  -dimension-name="": Dimension Name regex
  -dimension-value="": Dimension Value regex
  -dump=true: Dump raw metrics data received
  -fetch-previous=60: Negative time in minutes from now to fetch metrics
  -list=false: Print metrics available matching filter, but don't fetch
  -metrics="": Metrics name regex
  -namespace="": Namespace
  -period=1: Period (multiples of 60s)
  -region="us-east-1": AWS region
</pre>

### Examples

List all metrics available in the AWS/EC2 namespace:
<pre>
% cwfetch -namespace="AWS/EC2" -list
Dimensions: InstanceId=i-a138e34a, MetricName: NetworkIn
Dimensions: InstanceId=i-a138e34a, MetricName: StatusCheckFailed
Dimensions: InstanceId=i-613ae18a, MetricName: NetworkIn
Dimensions: InstanceId=i-b9a95598, MetricName: StatusCheckFailed_System
Dimensions: InstanceType=c3.xlarge, MetricName: CPUUtilization
Dimensions: InstanceId=i-26b5d25a, MetricName: DiskReadOps
Dimensions: InstanceId=i-26b5d25a, MetricName: NetworkIn
Dimensions: InstanceId=i-6487c58b, MetricName: CPUCreditUsage
...cont...
</pre>

List all metrics for instance i-a138e34a:
<pre>
% cwfetch -namespace="AWS/EC2" -dimension-value="i-a138e34a" -list
Dimensions: InstanceId=i-a138e34a, MetricName: StatusCheckFailed
Dimensions: InstanceId=i-a138e34a, MetricName: StatusCheckFailed_Instance
Dimensions: InstanceId=i-a138e34a, MetricName: DiskWriteBytes
Dimensions: InstanceId=i-a138e34a, MetricName: NetworkIn
Dimensions: InstanceId=i-a138e34a, MetricName: DiskReadBytes
Dimensions: InstanceId=i-a138e34a, MetricName: StatusCheckFailed_System
Dimensions: InstanceId=i-a138e34a, MetricName: NetworkOut
Dimensions: InstanceId=i-a138e34a, MetricName: CPUUtilization
Dimensions: InstanceId=i-a138e34a, MetricName: DiskWriteOps
Dimensions: InstanceId=i-a138e34a, MetricName: DiskReadOps
</pre>

List all network related metrics for instance i-a138e34a:
<pre>
% cwfetch -namespace="AWS/EC2" -dimension-value="i-a138e34a" -metrics="^Net" -list    
Dimensions: InstanceId=i-a138e34a, MetricName: NetworkIn
Dimensions: InstanceId=i-a138e34a, MetricName: NetworkOut
</pre>

Fetch last 2 hours of the referenced metrics:
<pre>
% cwfetch -namespace="AWS/EC2" -dimension-value="i-a138e34a" -metrics="^Net" -fetch-previous="120"
{"Datapoints":[{"Average":4.5039598e+06,"Maximum":null,"Minimum":null,"SampleCount":null, ...cont...
</pre>
