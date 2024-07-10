package cloudwatch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// ClientInterface for interacting with CloudWatch.
type ClientInterface interface {
	ListTagsForResource(ctx context.Context, params *cloudwatch.ListTagsForResourceInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.ListTagsForResourceOutput, error)
}

// MockClient used for testing purposes.
type MockClient struct {
	Tags []types.Tag
}

// ListTagsForResource mocks the CloudWatch API.
func (m *MockClient) ListTagsForResource(ctx context.Context, params *cloudwatch.ListTagsForResourceInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.ListTagsForResourceOutput, error) {
	return &cloudwatch.ListTagsForResourceOutput{
		Tags: m.Tags,
	}, nil
}
