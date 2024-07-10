package sts

import (
	"context"
	"encoding/base64"

	signerv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

const (
	clusterIDHeader  = "x-k8s-aws-id"
	v1Prefix         = "k8s-aws-v1."
	expireHeader     = "X-Amz-Expires"
	expireHeaderTime = "60"
)

// ClientInterface provides an interface for the STS client.
type ClientInterface interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(options *sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// PresignClientInterface provides an interface for the STS Presign client.
type PresignClientInterface interface {
	PresignGetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(options *sts.PresignOptions)) (*signerv4.PresignedHTTPRequest, error)
}

// TokenGenerator generates a token.
type TokenGenerator struct {
	stsClient PresignClientInterface
}

// NewTokenGenerator creates a new generator.
func NewTokenGenerator(stsClient PresignClientInterface) *TokenGenerator {
	return &TokenGenerator{
		stsClient: stsClient,
	}
}

// GenerateToken returns a token valid for clusterID using the given STS client.
func (g *TokenGenerator) GenerateToken(ctx context.Context, clusterID string) (string, error) {

	// This code is taken from https://github.com/kubernetes-sigs/aws-iam-authenticator/blob/master/pkg/token/token.go
	// and updated to use AWS SDK v2.

	// generate a sts:GetCallerIdentity request and add our custom cluster ID header
	request, err := g.stsClient.PresignGetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}, func(opts *sts.PresignOptions) {
		opts.ClientOptions = []func(*sts.Options){
			sts.WithAPIOptions(
				smithyhttp.AddHeaderValue(clusterIDHeader, clusterID),
				smithyhttp.AddHeaderValue(expireHeader, expireHeaderTime),
			),
		}
	})
	if err != nil {
		return "", err
	}

	token := v1Prefix + base64.RawURLEncoding.EncodeToString([]byte(request.URL))

	return token, err
}
