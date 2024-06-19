// Package lifecycle lements the lifecycle hooks
package lifecycle

import (
	"context"
	"fmt"

	"github.com/cloudnative-pg/cloudnative-pg/pkg/management/log"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper"
	"github.com/cloudnative-pg/cnpg-i/pkg/lifecycle"
	"github.com/dougkirkley/cnpg-plugin-s3-backup/internal/config"
	"github.com/dougkirkley/cnpg-plugin-s3-backup/internal/utils"

	"github.com/dougkirkley/cnpg-plugin-s3-backup/pkg/metadata"
)

// Lifecycle is the lementation of the lifecycle handler
type Lifecycle struct {
	lifecycle.UnimplementedOperatorLifecycleServer
}

// GetCapabilities exposes the lifecycle capabilities
func (l Lifecycle) GetCapabilities(
	_ context.Context,
	_ *lifecycle.OperatorLifecycleCapabilitiesRequest,
) (*lifecycle.OperatorLifecycleCapabilitiesResponse, error) {
	return &lifecycle.OperatorLifecycleCapabilitiesResponse{
		LifecycleCapabilities: []*lifecycle.OperatorLifecycleCapabilities{
			{
				Group: "",
				Kind:  "Pod",
				OperationTypes: []*lifecycle.OperatorOperationType{
					{
						Type: lifecycle.OperatorOperationType_TYPE_CREATE,
					},
					{
						Type: lifecycle.OperatorOperationType_TYPE_PATCH,
					},
				},
			},
		},
	}, nil
}

// LifecycleHook is called when creating Kubernetes services
func (l Lifecycle) LifecycleHook(
	ctx context.Context,
	request *lifecycle.OperatorLifecycleRequest,
) (*lifecycle.OperatorLifecycleResponse, error) {
	kind, err := utils.GetKind(request.ObjectDefinition)
	if err != nil {
		return nil, err
	}
	operation := request.OperationType.Type.Enum()
	if operation == nil {
		return nil, fmt.Errorf("no operation set")
	}

	switch kind {
	case "Pod":
		switch *operation {
		case lifecycle.OperatorOperationType_TYPE_CREATE:
			return l.reconcilePod(ctx, request)
		case lifecycle.OperatorOperationType_TYPE_UPDATE:
			return l.reconcilePod(ctx, request)
		}
	}

	return &lifecycle.OperatorLifecycleResponse{}, nil
}

// LifecycleHook is called when creating Kubernetes services
func (l Lifecycle) reconcilePod(
	ctx context.Context,
	request *lifecycle.OperatorLifecycleRequest,
) (*lifecycle.OperatorLifecycleResponse, error) {
	logger := log.FromContext(ctx).WithName("s3_backup_lifecycle")
	helper, err := pluginhelper.NewDataBuilder(
		metadata.PluginName,
		request.ClusterDefinition,
	).WithPod(request.ObjectDefinition).Build()
	if err != nil {
		return nil, err
	}
	configuration, valErrs := config.FromParameters(ctx, helper)
	if len(valErrs) > 0 {
		return nil, valErrs[0]
	}
	mutatedPod := helper.GetPod().DeepCopy()

	helper.InjectPluginVolume(mutatedPod)

	// Inject sidecar
	if len(mutatedPod.Spec.Containers) > 0 {
		mutatedPod.Spec.Containers = append(
			mutatedPod.Spec.Containers,
			getSidecarContainer(mutatedPod, helper.Parameters))
	}

	patch, err := helper.CreatePodJSONPatch(*mutatedPod)
	if err != nil {
		return nil, err
	}

	logger.Debug("generated patch", "content", string(patch), "configuration", configuration)

	return &lifecycle.OperatorLifecycleResponse{
		JsonPatch: patch,
	}, nil
}
