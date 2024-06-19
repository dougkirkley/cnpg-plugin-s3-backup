package operator

import (
	"context"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper"
	"github.com/cloudnative-pg/cnpg-i/pkg/operator"

	"github.com/dougkirkley/cnpg-plugin-s3-backup/internal/config"
	"github.com/dougkirkley/cnpg-plugin-s3-backup/pkg/metadata"
)

// ValidateClusterCreate validates a cluster that is being created
func (Operator) ValidateClusterCreate(
	ctx context.Context,
	request *operator.OperatorValidateClusterCreateRequest,
) (*operator.OperatorValidateClusterCreateResult, error) {
	result := &operator.OperatorValidateClusterCreateResult{}

	helper, err := pluginhelper.NewDataBuilder(
		metadata.PluginName,
		request.Definition,
	).Build()
	if err != nil {
		return nil, err
	}

	_, result.ValidationErrors = config.FromParameters(ctx, helper)

	return result, nil
}

// ValidateClusterChange validates a cluster that is being changed
func (Operator) ValidateClusterChange(
	ctx context.Context,
	request *operator.OperatorValidateClusterChangeRequest,
) (*operator.OperatorValidateClusterChangeResult, error) {
	result := &operator.OperatorValidateClusterChangeResult{}

	return result, nil
}
