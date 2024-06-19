package config

import (
	"context"

	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper"
	"github.com/cloudnative-pg/cnpg-i/pkg/operator"
)

const (
	ImageNameParam       = "image"
	ImagePullPolicyParam = "imagePullPolicy"
	RegionParam          = "region"
	EndpointParam        = "endpoint"
	AwsKeyParam          = "aws_key"
	AwsSecretKeyParam    = "aws_secret_key"
	BucketParam          = "bucket"
	PrefixParam          = "prefix"
)

// Configuration represents the plugin configuration parameters
type Configuration struct {
	Image           string
	ImagePullPolicy string
	Region          string
	Endpoint        string
	AwsKey          string
	AwsSecretKey    string
	Bucket          string
	Prefix          string
}

// FromParameters builds a plugin configuration from the configuration parameters
func FromParameters(
	_ context.Context,
	helper *pluginhelper.Data,
) (*Configuration, []*operator.ValidationError) {
	validationErrors := make([]*operator.ValidationError, 0)

	if len(helper.Parameters[ImageNameParam]) == 0 {
		validationErrors = append(
			validationErrors,
			helper.ValidationErrorForParameter(ImageNameParam, "image cannot be empty"),
		)
	}

	configuration := &Configuration{
		Image:           helper.Parameters[ImageNameParam],
		ImagePullPolicy: helper.Parameters[ImagePullPolicyParam],
		Region:          helper.Parameters[RegionParam],
		Endpoint:        helper.Parameters[EndpointParam],
		AwsKey:          helper.Parameters[AwsKeyParam],
		AwsSecretKey:    helper.Parameters[AwsSecretKeyParam],
		Bucket:          helper.Parameters[BucketParam],
		Prefix:          helper.Parameters[PrefixParam],
	}

	return configuration, validationErrors
}

// ToParameters serialize the configuration to a map of plugin parameters
func (config *Configuration) ToParameters() (map[string]string, error) {
	result := map[string]string{
		ImageNameParam:       config.Image,
		ImagePullPolicyParam: config.ImagePullPolicy,
		RegionParam:          config.Region,
		EndpointParam:        config.Endpoint,
		AwsKeyParam:          config.AwsKey,
		AwsSecretKeyParam:    config.AwsSecretKey,
		BucketParam:          config.Bucket,
		PrefixParam:          config.Prefix,
	}

	return result, nil
}
