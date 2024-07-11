package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awscloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/skpr/lambda-eks-event-cloudwatch/internal/cloudwatch"
	skpreks "github.com/skpr/lambda-eks-event-cloudwatch/internal/eks"
	"github.com/skpr/lambda-eks-event-cloudwatch/pkg/annotation"
	skpraws "github.com/skpr/lambda-eks-event-cloudwatch/pkg/aws"
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

	log.Printf("Validating event")

	if event.AlarmARN == "" {
		return fmt.Errorf("alarm ARN is required")
	}

	if event.AlarmData.Configuration.Description == "" {
		return fmt.Errorf("alarm configuration description is required")
	}

	log.Printf("Looking up alarm tags")

	alarm, err := awscloudwatch.NewFromConfig(cfg).ListTagsForResource(ctx, &awscloudwatch.ListTagsForResourceInput{
		ResourceARN: aws.String(event.AlarmARN),
	})
	if err != nil {
		return fmt.Errorf("failed to list tags for resource: %w", err)
	}

	cluster, err := cloudwatch.GetValueFromTag(alarm.Tags, skpraws.TagKeyCluster)
	if err != nil {
		return fmt.Errorf("failed to get cluster from tags: %w", err)
	}

	apiGroup, err := cloudwatch.GetValueFromTag(alarm.Tags, skpraws.TagKeyAPIGroup)
	if err != nil {
		return fmt.Errorf("failed to get api version from tags: %w", err)
	}

	apiVersion, err := cloudwatch.GetValueFromTag(alarm.Tags, skpraws.TagKeyAPIVersion)
	if err != nil {
		return fmt.Errorf("failed to get api version from tags: %w", err)
	}

	kind, err := cloudwatch.GetValueFromTag(alarm.Tags, skpraws.TagKeyKind)
	if err != nil {
		return fmt.Errorf("failed to get kind from tags: %w", err)
	}

	namespace, err := cloudwatch.GetValueFromTag(alarm.Tags, skpraws.TagKeyNamespace)
	if err != nil {
		return fmt.Errorf("failed to get namespace from tags: %w", err)
	}

	name, err := cloudwatch.GetValueFromTag(alarm.Tags, skpraws.TagKeyName)
	if err != nil {
		return fmt.Errorf("failed to get name from tags: %w", err)
	}

	reason, err := cloudwatch.GetValueFromTag(alarm.Tags, skpraws.TagKeyReason)
	if err != nil {
		return fmt.Errorf("failed to get reason from tags: %w", err)
	}

	log.Printf("Connecting to EKS cluster")

	config, err := skpreks.BuildKubeconfig(ctx, eks.NewFromConfig(cfg), cluster)
	if err != nil {
		return fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to get kubernetes clientset: %w", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	log.Printf("Looking up resource version and UID")

	res := schema.GroupVersionResource{Group: "workflow.skpr.io", Version: "v1beta1", Resource: "environments"}

	unstructured, err := client.Resource(res).Namespace("skpr-project-drupal").Get(context.TODO(), "prod", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	log.Printf("Marshalling to Kubernetes event")

	object := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: "aws-cloudwatch-alarm-",
			Annotations: map[string]string{
				annotation.KeyCloudWatchAlarmName: event.AlarmData.AlarmName,
			},
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion:      fmt.Sprintf("%s/%s", apiGroup, apiVersion),
			Kind:            kind,
			Namespace:       namespace,
			Name:            name,
			UID:             unstructured.GetUID(),
			ResourceVersion: unstructured.GetResourceVersion(),
		},
		Type:           corev1.EventTypeWarning,
		Reason:         reason,
		Message:        event.AlarmData.Configuration.Description,
		FirstTimestamp: metav1.Now(),
		LastTimestamp:  metav1.Now(),
		Source: corev1.EventSource{
			Component: "aws-cloudwatch-alarm",
		},
	}

	log.Printf("Creating event")

	_, err = clientset.CoreV1().Events(object.ObjectMeta.Namespace).Create(context.TODO(), object, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	log.Println("Function complete")

	return nil
}
