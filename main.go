package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awscloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/skpr/lambda-eks-event-cloudwatch/internal/cloudwatch"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/skpr/lambda-eks-event-cloudwatch/pkg/annotation"
)

var (
	// GitVersion overridden at build time by:
	//   -ldflags="-X main.GitVersion=${VERSION}"
	GitVersion string
)

func main() {
	lambda.Start(HandleLambdaEvent)
}

// HandleLambdaEvent will respond to a CloudWatch Alarm, check for rate limited IP addresses and send a message to Slack.
func HandleLambdaEvent(ctx context.Context, event *cloudwatch.Event) error {
	log.Printf("Running Lambda (%s)\n", GitVersion)

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	cloudwatchClient := awscloudwatch.NewFromConfig(cfg)

	err = run(ctx, cloudwatchClient, event)
	if err != nil {
		return err
	}

	log.Println("Function complete")

	return nil
}

// Run will execute the core of the function.
func run(ctx context.Context, cloudwatchClient cloudwatch.ClientInterface, event *cloudwatch.Event) error {
	if event.AlarmARN == "" {
		return fmt.Errorf("alarm ARN is required")
	}

	if event.AlarmData.State.Reason == "" {
		return fmt.Errorf("alarm state reason is required")
	}

	if event.AlarmData.Configuration.Description == "" {
		return fmt.Errorf("alarm configuration description is required")
	}

	alarm, err := cloudwatchClient.ListTagsForResource(ctx, &awscloudwatch.ListTagsForResourceInput{
		ResourceARN: aws.String(event.AlarmARN),
	})
	if err != nil {
		return fmt.Errorf("failed to list tags for resource: %w", err)
	}

	apiVersion, err := cloudwatch.GetValueFromTag(alarm.Tags, cloudwatch.TagKeyAPIVersion)
	if err != nil {
		return fmt.Errorf("failed to get api version from tags: %w", err)
	}

	kind, err := cloudwatch.GetValueFromTag(alarm.Tags, cloudwatch.TagKeyKind)
	if err != nil {
		return fmt.Errorf("failed to get kind from tags: %w", err)
	}

	namespace, err := cloudwatch.GetValueFromTag(alarm.Tags, cloudwatch.TagKeyNamespace)
	if err != nil {
		return fmt.Errorf("failed to get namespace from tags: %w", err)
	}

	name, err := cloudwatch.GetValueFromTag(alarm.Tags, cloudwatch.TagKeyName)
	if err != nil {
		return fmt.Errorf("failed to get name from tags: %w", err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		return fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to get kubernetes clientset: %w", err)
	}

	e := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: "aws-cloudwatch-alarm-",
			Annotations: map[string]string{
				annotation.KeyCloudWatchAlarmName: event.AlarmData.AlarmName,
			},
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion: apiVersion,
			Kind:       kind,
			Namespace:  namespace,
			Name:       name,
		},
		Type:    corev1.EventTypeWarning,
		Reason:  event.AlarmData.State.Reason,
		Message: event.AlarmData.Configuration.Description,
	}

	_, err = clientset.CoreV1().Events(namespace).Create(context.TODO(), e, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}
