package operator

import (
	"context"

	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper"
	"github.com/cloudnative-pg/cnpg-i/pkg/operator"

	"github.com/dougkirkley/cnpg-plugin-s3-backup/pkg/metadata"

	"github.com/dougkirkley/cnpg-plugin-s3-backup/internal/config"
)

// MutateCluster is called to mutate a cluster with the defaulting webhook.
// This function is defaulting the "imagePullPolicy" plugin parameter
func (Operator) MutateCluster(
	ctx context.Context,
	request *operator.OperatorMutateClusterRequest,
) (*operator.OperatorMutateClusterResult, error) {
	helper, err := pluginhelper.NewDataBuilder(
		metadata.PluginName,
		request.Definition,
	).Build()
	if err != nil {
		return nil, err
	}

	cfg, valErrs := config.FromParameters(ctx, helper)
	if len(valErrs) > 0 {
		return nil, valErrs[0]
	}

	mutatedCluster := helper.GetCluster().DeepCopy()
	for i := range mutatedCluster.Spec.Plugins {
		if mutatedCluster.Spec.Plugins[i].Name != metadata.PluginName {
			continue
		}

		if mutatedCluster.Spec.Plugins[i].Parameters == nil {
			mutatedCluster.Spec.Plugins[i].Parameters = make(map[string]string)
		}

		mutatedCluster.Spec.Plugins[i].Parameters, err = cfg.ToParameters()
		if err != nil {
			return nil, err
		}
	}

	patch, err := helper.CreateClusterJSONPatch(*mutatedCluster)
	if err != nil {
		return nil, err
	}

	return &operator.OperatorMutateClusterResult{
		JsonPatch: patch,
	}, nil
}
