package buildpack

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func rewriteLayer(layer v1.Layer, oldID, newID, oldVersion, newVersion string) (v1.Layer, error) {
	b := &bytes.Buffer{}
	tw := tar.NewWriter(b)

	uncompressed, err := layer.Uncompressed()
	if err != nil {
		return nil, fmt.Errorf("unable to get uncompressed layer contents\n%w", err)
	}
	defer uncompressed.Close()

	tr := tar.NewReader(uncompressed)
	for {
		header, err := tr.Next()
		if err != nil {
			break
		}

		// replace buildpack id and version in folder names
		newName := strings.ReplaceAll(
			strings.ReplaceAll(header.Name, escapedID(oldID), escapedID(newID)),
			oldVersion, newVersion)

		// replace buildpack id and version in buildpack.toml
		if strings.HasSuffix(path.Clean(header.Name), "buildpack.toml") {
			buf, err := ioutil.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("unable to read buildpack.toml\n%w", err)
			}

			bd := BuildpackDescriptor{}
			_, err = toml.Decode(string(buf), &bd)
			if err != nil {
				return nil, fmt.Errorf("unable to decode buildpack.toml\n%w", err)
			}

			bd.Info.ID = newID
			bd.Info.Version = newVersion

			updatedBuildpackToml := &bytes.Buffer{}
			err = toml.NewEncoder(updatedBuildpackToml).Encode(bd)
			if err != nil {
				return nil, fmt.Errorf("unable to encode buildpack.toml\n%w", err)
			}

			contents := updatedBuildpackToml.Bytes()
			header.Name = newName
			header.Size = int64(len(contents))
			err = tw.WriteHeader(header)
			if err != nil {
				return nil, fmt.Errorf("unable to write updated header\n%w", err)
			}

			_, err = tw.Write(contents)
			if err != nil {
				return nil, fmt.Errorf("unable to write updated contents\n%w", err)
			}
		} else {
			header.Name = newName
			err = tw.WriteHeader(header)
			if err != nil {
				return nil, fmt.Errorf("unable to write header\n%w", err)
			}

			buf, err := ioutil.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("unable to read contents\n%w", err)
			}

			_, err = tw.Write(buf)
			if err != nil {
				return nil, fmt.Errorf("unable to write contents\n%w", err)
			}
		}
	}

	return tarball.LayerFromReader(b)
}

type BuildpackDescriptor struct {
	API      string            `toml:"api"`
	Info     BuildpackTomlInfo `toml:"buildpack"`
	Stacks   interface{}       `toml:"stacks"`
	Order    interface{}       `toml:"order"`
	Metadata interface{}       `toml:"metadata"`
}

type BuildpackTomlInfo struct {
	ID       string `toml:"id"`
	Version  string `toml:"version"`
	Name     string `toml:"name"`
	ClearEnv bool   `toml:"clear-env,omitempty"`
}

func escapedID(id string) string {
	return strings.ReplaceAll(id, "/", "_")
}
