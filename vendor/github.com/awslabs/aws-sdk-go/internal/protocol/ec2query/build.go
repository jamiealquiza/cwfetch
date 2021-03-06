package ec2query

//go:generate go run ../../fixtures/protocol/generate.go ../../fixtures/protocol/input/ec2.json build_test.go

import (
	"net/url"

	"github.com/jamiealquiza/cloudwatch-graphite/vendor/github.com/awslabs/aws-sdk-go/aws"
	"github.com/jamiealquiza/cloudwatch-graphite/vendor/github.com/awslabs/aws-sdk-go/internal/protocol/query/queryutil"
)

func Build(r *aws.Request) {
	body := url.Values{
		"Action":  {r.Operation.Name},
		"Version": {r.Service.APIVersion},
	}
	if err := queryutil.Parse(body, r.Params, true); err != nil {
		r.Error = err
		return
	}

	r.HTTPRequest.Method = "POST"
	r.HTTPRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	r.SetBufferBody([]byte(body.Encode()))
}
