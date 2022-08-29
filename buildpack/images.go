package buildpack

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const (
	layerMetadataLabel     = "io.buildpacks.buildpack.layers"
	buildpackMetadataLabel = "io.buildpacks.buildpackage.metadata"
)

func Rename(buildpack, tag, newID, newVersion string) (string, error) {
	reference, err := name.ParseReference(buildpack)
	if err != nil {
		return "", fmt.Errorf("unable to parse reference for existing buildpack tag\n%w", err)
	}

	image, err := remote.Image(reference, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", fmt.Errorf("unable to fetch remote image\n%w", err)
	}

	metadata := Metadata{}
	err = GetLabel(image, buildpackMetadataLabel, &metadata)
	if err != nil {
		return "", fmt.Errorf("unable to get buildpack metadata label\n%w", err)
	}

	layerMetadata := BuildpackLayerMetadata{}
	err = GetLabel(image, layerMetadataLabel, &layerMetadata)
	if err != nil {
		return "", fmt.Errorf("unable to get buildpack layer metadata\n%w", err)
	}

	newLayersMetedata, layers, err := layerMetadata.metadataAndLayersFor(image, metadata.Id, metadata.Version, newID, newVersion)
	if err != nil {
		return "", fmt.Errorf("unable to generate new metadata\n%w", err)
	}

	newBuildpackage, err := random.Image(0, 0)
	if err != nil {
		return "", fmt.Errorf("unable to create new image\n%w", err)
	}

	newBuildpackage, err = mutate.AppendLayers(newBuildpackage, layers...)
	if err != nil {
		return "", fmt.Errorf("unable to append layers to new image\n%w", err)
	}

	metadata.Id = newID
	metadata.Version = newVersion
	newBuildpackage, err = SetLabels(newBuildpackage, map[string]interface{}{
		layerMetadataLabel:     newLayersMetedata,
		buildpackMetadataLabel: metadata,
	})
	if err != nil {
		return "", fmt.Errorf("unable to set buildpack layer metadata label\n%w", err)
	}

	reference, err = name.ParseReference(tag)
	if err != nil {
		return "", fmt.Errorf("unable to unable to parse reference for new buildpack tag\n%w", err)
	}

	srcCfgFile, err := image.ConfigFile()
	if err != nil {
		return "", fmt.Errorf("unable to fetch config file\n%w", err)
	}

	targetCfgFile, err := newBuildpackage.ConfigFile()
	if err != nil {
		return "", fmt.Errorf("unable to fetch config file\n%w", err)
	}
	targetCfgFile.OS = srcCfgFile.OS

	newBuildpackage, err = mutate.ConfigFile(newBuildpackage, targetCfgFile)
	if err != nil {
		return "", fmt.Errorf("unable to transfer config file\n%w", err)
	}

	err = remote.Write(reference, newBuildpackage, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", fmt.Errorf("unable to write new buildapck\n%w", err)
	}

	digest, err := newBuildpackage.Digest()
	if err != nil {
		return "", fmt.Errorf("unable to calculate new buildpack digest\n%w", err)
	}

	identifer := fmt.Sprintf("%s@%s", tag, digest.String())

	return identifer, nil
}

func (m BuildpackLayerMetadata) metadataAndLayersFor(sourceImage v1.Image, oldId string, oldVersion string, newId string, newVersion string) (BuildpackLayerMetadata, []v1.Layer, error) {
	newLayerMetdata := BuildpackLayerMetadata{}

	var layers []v1.Layer
	for id, versions := range m {
		for v, buildpack := range versions {

			if v != oldVersion || id != oldId {
				if _, ok := newLayerMetdata[id]; !ok {
					newLayerMetdata[id] = map[string]BuildpackLayerInfo{}
				}

				newLayerMetdata[id][v] = buildpack

				diffId, err := v1.NewHash(buildpack.LayerDiffID)
				if err != nil {
					return nil, nil, fmt.Errorf("unable to create new diff hash %s for non-matching layer\n%w", buildpack.LayerDiffID, err)
				}
				layer, err := sourceImage.LayerByDiffID(diffId)
				if err != nil {
					return nil, nil, fmt.Errorf("unable to fetch layer by diff id %s for non-matching layer\n%w", diffId, err)
				}

				layers = append(layers, layer)
			} else {
				if _, ok := newLayerMetdata[newId]; !ok {
					newLayerMetdata[newId] = map[string]BuildpackLayerInfo{}
				}

				diffId, err := v1.NewHash(buildpack.LayerDiffID)
				if err != nil {
					return nil, nil, fmt.Errorf("unable to create new diff hash %s for matching layer\n%w", buildpack.LayerDiffID, err)
				}
				layer, err := sourceImage.LayerByDiffID(diffId)
				if err != nil {
					return nil, nil, fmt.Errorf("unable to fetch layer by diff id %s for matching layer\n%w", diffId, err)
				}

				layer, err = rewriteLayer(layer, oldId, newId, oldVersion, newVersion)
				if err != nil {
					return nil, nil, fmt.Errorf("unable to rewrite layer, old id: %s, new id: %s, new version: %s\n%w", oldId, newId, newVersion, err)
				}

				diffID, err := layer.DiffID()
				if err != nil {
					return nil, nil, fmt.Errorf("unable to get hash of layer\n%w", err)
				}

				buildpack.LayerDiffID = diffID.String()

				newLayerMetdata[newId][newVersion] = buildpack
				layers = append(layers, layer)
			}
		}
	}

	return newLayerMetdata, layers, nil
}

type BuildpackLayerMetadata map[string]map[string]BuildpackLayerInfo

type BuildpackLayerInfo struct {
	API         string      `json:"api"`
	Stacks      []Stack     `json:"stacks,omitempty"`
	Order       Order       `json:"order,omitempty"`
	LayerDiffID string      `json:"layerDiffID"`
	Name        string      `json:"name,omitempty"`
	Homepage    interface{} `json:"homepage,omitempty"`
}

type Order []OrderEntry

type OrderEntry struct {
	Group []BuildpackRef `json:"group,omitempty"`
}

type BuildpackRef struct {
	BuildpackInfo `json:",inline"`
	Optional      bool `json:"optional,omitempty"`
}

type BuildpackInfo struct {
	Id      string `json:"id"`
	Version string `json:"version,omitempty"`
}

type Stack struct {
	ID     string   `json:"id"`
	Mixins []string `json:"mixins,omitempty"`
}

type Metadata struct {
	Id          string      `json:"id"`
	Name        string      `json:"name,omitempty"`
	Version     string      `json:"version,omitempty"`
	Homepage    interface{} `json:"homepage,omitempty"`
	Description interface{} `json:"description,omitempty"`
	Keywords    interface{} `json:"keywords,omitempty"`
	Licenses    interface{} `json:"licenses,omitempty"`
	Stacks      interface{} `toml:"stacks" json:"stacks,omitempty"`
}
