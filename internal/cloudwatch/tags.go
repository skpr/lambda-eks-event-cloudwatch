package cloudwatch

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
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
