package cloudwatch

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

const (
	// TagKeyAPIVersion is used to determine the API version of Kubernetes resource.
	TagKeyAPIVersion = "skpr.io/k8s-event-api-version"
	// TagKeyKind is used to determine the kind of Kubernetes resource.
	TagKeyKind = "skpr.io/k8s-event-kind"
	// TagKeyCluster is used to determine the cluster where we will send events.
	TagKeyCluster = "skpr.io/k8s-event-cluster"
	// TagKeyNamespace is used to determine the namespace of the Kubernetes resource.
	TagKeyNamespace = "skpr.io/k8s-event-namespace"
	// TagKeyName is used to determine the name of the Kubernetes resource.
	TagKeyName = "skpr.io/k8s-event-name"
	// TagKeyReason is used to determine the reason for this event.
	TagKeyReason = "skpr.io/k8s-event-reason"
)

// GetValueFromTag returns the value of a tag with the given key, or false if the tag does not exist.
func GetValueFromTag(tags []types.Tag, key string) (string, error) {
	for _, tag := range tags {
		if *tag.Key == key {
			return *tag.Value, nil
		}
	}

	return "", fmt.Errorf("tag not found")
}
