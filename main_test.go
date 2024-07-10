package main

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscloudwatchtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/skpr/lambda-eks-event-cloudwatch/internal/cloudwatch"
	"github.com/skpr/lambda-eks-event-cloudwatch/pkg/annotation"
)

func TestGetKubernetesEvent(t *testing.T) {
	tags := []awscloudwatchtypes.Tag{
		{
			Key:   aws.String(cloudwatch.TagKeyAPIVersion),
			Value: aws.String("test.skpr.io/v1beta1"),
		},
		{
			Key:   aws.String(cloudwatch.TagKeyKind),
			Value: aws.String("Test"),
		},
		{
			Key:   aws.String(cloudwatch.TagKeyCluster),
			Value: aws.String("test-cluster"),
		},
		{
			Key:   aws.String(cloudwatch.TagKeyNamespace),
			Value: aws.String("test-namespace"),
		},
		{
			Key:   aws.String(cloudwatch.TagKeyName),
			Value: aws.String("test-object"),
		},
	}

	event := &cloudwatch.Event{
		AlarmData: cloudwatch.AlarmData{
			AlarmName: "test-alarm",
			State: cloudwatch.AlarmDataState{
				Reason: "test-reason",
			},
			Configuration: cloudwatch.AlarmDataConfiguration{
				Description: "test-description",
			},
		},
	}

	object, err := getKubernetesEvent(tags, event)
	assert.NoError(t, err)

	want := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "test-namespace",
			GenerateName: "aws-cloudwatch-alarm-",
			Annotations: map[string]string{
				annotation.KeyCloudWatchAlarmName: "test-alarm",
			},
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion: "test.skpr.io/v1beta1",
			Kind:       "Test",
			Namespace:  "test-namespace",
			Name:       "test-object",
		},
		Type:    corev1.EventTypeWarning,
		Reason:  event.AlarmData.State.Reason,
		Message: event.AlarmData.Configuration.Description,
	}

	assert.Equal(t, want, object)
}
