package lifecycle

import (
	"github.com/dougkirkley/cnpg-plugin-s3-backup/internal/config"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const pgPath = "/var/lib/postgresql"

func getSidecarContainer(pgPod *corev1.Pod, parameters map[string]string) corev1.Container {
	result := corev1.Container{
		Name: "plugin-s3-backup",
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "scratch-data",
				MountPath: "/controller",
			},
			{
				Name:      "plugins",
				MountPath: "/plugins",
			},
		},
		Image:           parameters[config.ImageNameParam],
		ImagePullPolicy: corev1.PullPolicy(parameters[config.ImagePullPolicyParam]),
		Env: []corev1.EnvVar{
			{
				Name:  "",
				Value: parameters[config.PrefixParam],
			},
		},
	}

	volumeMounts := pgPod.Spec.Containers[0].VolumeMounts
	for i := range volumeMounts {
		if strings.HasPrefix(volumeMounts[i].MountPath, pgPath) {
			result.VolumeMounts = append(result.VolumeMounts, volumeMounts[i])
		}
	}

	if len(parameters[config.BucketParam]) > 0 {
		result.Env = append(result.Env, corev1.EnvVar{
			Name:  "AWS_BUCKET",
			Value: parameters[config.BucketParam],
		})
	}

	if len(parameters[config.BucketParam]) > 0 {
		result.Env = append(result.Env, corev1.EnvVar{
			Name:  "BACKUP_PREFIX",
			Value: parameters[config.PrefixParam],
		})
	}

	if len(parameters[config.RegionParam]) > 0 {
		result.Env = append(result.Env, corev1.EnvVar{
			Name:  "AWS_REGION",
			Value: parameters[config.RegionParam],
		})
	}

	if len(parameters[config.EndpointParam]) > 0 {
		result.Env = append(result.Env, corev1.EnvVar{
			Name:  "AWS_ENDPOINT_URL",
			Value: parameters[config.EndpointParam],
		})
	}

	if len(parameters[config.AwsKeyParam]) > 0 {
		result.Env = append(result.Env, corev1.EnvVar{
			Name:  "AWS_ACCESS_KEY_ID",
			Value: parameters[config.AwsKeyParam],
		})
	}

	if len(parameters[config.AwsSecretKeyParam]) > 0 {
		result.Env = append(result.Env, corev1.EnvVar{
			Name:  "AWS_SECRET_ACCESS_KEY",
			Value: parameters[config.AwsSecretKeyParam],
		})
	}

	return result
}
