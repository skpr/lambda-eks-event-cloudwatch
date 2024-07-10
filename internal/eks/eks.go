package eks

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"k8s.io/client-go/rest"

	skprsts "github.com/skpr/lambda-eks-event-cloudwatch/internal/sts"
)

// BuildKubeconfig for a given EKS cluster.
func BuildKubeconfig(ctx context.Context, eksClient *eks.Client, cluster string) (*rest.Config, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get aws config: %w", err)
	}

	// Query EKS for the CA etc.
	resp, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(cluster),
	})
	if err != nil {
		return nil, err
	}

	ca, err := base64.StdEncoding.DecodeString(*resp.Cluster.CertificateAuthority.Data)
	if err != nil {
		return nil, err
	}

	var (
		stsClient        = sts.NewFromConfig(cfg)
		stsPresignClient = sts.NewPresignClient(stsClient)
	)

	gen := skprsts.NewTokenGenerator(stsPresignClient)

	token, err := gen.GenerateToken(ctx, cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to get sts token: %w", err)
	}

	return &rest.Config{
		UserAgent:   "Skpr Lambda EKS Event CloudWatch",
		BearerToken: token,
		Host:        *resp.Cluster.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	}, nil
}
