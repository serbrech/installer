package cluster

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/cluster/aws"
	"github.com/openshift/installer/pkg/asset/cluster/azure"
	"github.com/openshift/installer/pkg/asset/cluster/libvirt"
	"github.com/openshift/installer/pkg/asset/cluster/openstack"
	"github.com/openshift/installer/pkg/asset/installconfig"
	"github.com/openshift/installer/pkg/types"
	awstypes "github.com/openshift/installer/pkg/types/aws"
	azuretypes "github.com/openshift/installer/pkg/types/azure"
	libvirttypes "github.com/openshift/installer/pkg/types/libvirt"
	nonetypes "github.com/openshift/installer/pkg/types/none"
	openstacktypes "github.com/openshift/installer/pkg/types/openstack"
	vspheretypes "github.com/openshift/installer/pkg/types/vsphere"
	"github.com/pkg/errors"
)

const (
	metadataFileName = "metadata.json"
)

// Metadata contains information needed to destroy clusters.
type Metadata struct {
	file *asset.File
}

var _ asset.WritableAsset = (*Metadata)(nil)

// Name returns the human-friendly name of the asset.
func (m *Metadata) Name() string {
	return "Metadata"
}

// Dependencies returns the direct dependencies for the metadata
// asset.
func (m *Metadata) Dependencies() []asset.Asset {
	return []asset.Asset{
		&installconfig.ClusterID{},
		&installconfig.InstallConfig{},
	}
}

// Generate generates the metadata asset.
func (m *Metadata) Generate(parents asset.Parents) (err error) {
	clusterID := &installconfig.ClusterID{}
	installConfig := &installconfig.InstallConfig{}
	parents.Get(clusterID, installConfig)

	metadata := &types.ClusterMetadata{
		ClusterName: installConfig.Config.ObjectMeta.Name,
		ClusterID:   clusterID.UUID,
		InfraID:     clusterID.InfraID,
	}

	switch installConfig.Config.Platform.Name() {
	case awstypes.Name:
		metadata.ClusterPlatformMetadata.AWS = aws.Metadata(clusterID.UUID, clusterID.InfraID, installConfig.Config)
	case libvirttypes.Name:
		metadata.ClusterPlatformMetadata.Libvirt = libvirt.Metadata(installConfig.Config)
	case openstacktypes.Name:
		metadata.ClusterPlatformMetadata.OpenStack = openstack.Metadata(clusterID.InfraID, installConfig.Config)
	case nonetypes.Name, vspheretypes.Name:
	case azuretypes.Name:
		metadata.ClusterPlatformMetadata.Azure = azure.Metadata(clusterID.UUID, clusterID.InfraID, installConfig.Config)
	default:
		return errors.Errorf("no known platform")
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return errors.Wrap(err, "failed to Marshal ClusterMetadata")
	}

	m.file = &asset.File{
		Filename: metadataFileName,
		Data:     data,
	}

	return nil
}

// Files returns the FileList generated by the asset.
func (m *Metadata) Files() []*asset.File {
	if m.file != nil {
		return []*asset.File{m.file}
	}
	return []*asset.File{}
}

// Load is a no-op, because we never want to load broken metadata from
// the disk.
func (m *Metadata) Load(f asset.FileFetcher) (found bool, err error) {
	return false, nil
}

// LoadMetadata loads the cluster metadata from an asset directory.
func LoadMetadata(dir string) (*types.ClusterMetadata, error) {
	path := filepath.Join(dir, metadataFileName)
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata *types.ClusterMetadata
	if err = json.Unmarshal(raw, &metadata); err != nil {
		return nil, errors.Wrapf(err, "failed to Unmarshal data from %q to types.ClusterMetadata", path)
	}

	return metadata, err
}
