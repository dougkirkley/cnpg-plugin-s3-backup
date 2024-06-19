package identity

import (
	"context"

	"github.com/cloudnative-pg/cnpg-i/pkg/identity"

	"github.com/dougkirkley/cnpg-plugin-s3-backup/pkg/metadata"
)

// Identity is the implementation of the identity service
type Identity struct {
	identity.IdentityServer
}

// GetPluginMetadata implements the IdentityServer interface
func (Identity) GetPluginMetadata(
	context.Context,
	*identity.GetPluginMetadataRequest,
) (*identity.GetPluginMetadataResponse, error) {
	return &metadata.Data, nil
}

// GetPluginCapabilities implements the IdentityServer interface
func (Identity) GetPluginCapabilities(
	context.Context,
	*identity.GetPluginCapabilitiesRequest,
) (*identity.GetPluginCapabilitiesResponse, error) {
	return &identity.GetPluginCapabilitiesResponse{
		Capabilities: []*identity.PluginCapability{
			{
				Type: &identity.PluginCapability_Service_{
					Service: &identity.PluginCapability_Service{
						Type: identity.PluginCapability_Service_TYPE_OPERATOR_SERVICE,
					},
				},
			},
			{
				Type: &identity.PluginCapability_Service_{
					Service: &identity.PluginCapability_Service{
						Type: identity.PluginCapability_Service_TYPE_BACKUP_SERVICE,
					},
				},
			},
			{
				Type: &identity.PluginCapability_Service_{
					Service: &identity.PluginCapability_Service{
						Type: identity.PluginCapability_Service_TYPE_LIFECYCLE_SERVICE,
					},
				},
			},
		},
	}, nil
}

// Probe implements the IdentityServer interface
func (Identity) Probe(context.Context, *identity.ProbeRequest) (*identity.ProbeResponse, error) {
	return &identity.ProbeResponse{
		Ready: true,
	}, nil
}
