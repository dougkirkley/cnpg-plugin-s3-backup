// Package metadata contains the metadata of this plugin
package metadata

import "github.com/cloudnative-pg/cnpg-i/pkg/identity"

// PluginName is the name of the plugin
const PluginName = "s3-backup.cloudnative-pg.io"

// Data is the metadata of this plugin
var Data = identity.GetPluginMetadataResponse{
	Name:          PluginName,
	Version:       "0.1.0",
	DisplayName:   "CNPG s3 backup plugin",
	ProjectUrl:    "https://github.com/dougkirkley/cnpg-plugin-s3-backup",
	RepositoryUrl: "https://github.com/dougkirkley/cnpg-plugin-s3-backup",
	License:       "Apache 2.0",
	Maturity:      "alpha",
}
