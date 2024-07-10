package sts

import (
	"context"
	"testing"

	signerv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
)

// PresignClient is a mock client.
type PresignClient struct {
	PresignClientInterface
}

// PresignGetCallerIdentity implements the interface.
func (c *PresignClient) PresignGetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(options *sts.PresignOptions)) (*signerv4.PresignedHTTPRequest, error) {
	return &signerv4.PresignedHTTPRequest{
		URL: "http://example/com",
	}, nil
}

func TestGenerateToken(t *testing.T) {
	stsClient := &PresignClient{}
	generator := NewTokenGenerator(stsClient)
	token, err := generator.GenerateToken(context.TODO(), "foo")
	assert.NoError(t, err)
	assert.Equal(t, "k8s-aws-v1.aHR0cDovL2V4YW1wbGUvY29t", token)
}
