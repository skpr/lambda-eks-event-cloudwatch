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
	skpraws "github.com/skpr/lambda-eks-event-cloudwatch/pkg/aws"
)

func TestGetKubernetesEvent(t *testing.T) {
	tags := []awscloudwatchtypes.Tag{
		{
			Key:   aws.String(skpraws.TagKeyAPIVersion),
			Value: aws.String("test.skpr.io/v1beta1"),
		},
		{
			Key:   aws.String(skpraws.TagKeyKind),
			Value: aws.String("Test"),
		},
		{
			Key:   aws.String(skpraws.TagKeyCluster),
			Value: aws.String("test-cluster"),
		},
		{
			Key:   aws.String(skpraws.TagKeyNamespace),
			Value: aws.String("test-namespace"),
		},
		{
			Key:   aws.String(skpraws.TagKeyName),
			Value: aws.String("test-object"),
		},
		{
			Key:   aws.String(skpraws.TagKeyReason),
			Value: aws.String("test-reason"),
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
		Reason:  "test-reason",
		Message: "test-description",
	}

	assert.Equal(t, want.ObjectMeta, object.ObjectMeta, "ObjectMeta matches")
	assert.Equal(t, want.InvolvedObject, object.InvolvedObject, "InvolvedObject matches")
	assert.Equal(t, want.Type, object.Type, "Type matches")
	assert.Equal(t, want.Reason, object.Reason, "Reason matches")
	assert.Equal(t, want.Message, object.Message, "Message matches")
}
