/*
 * Copyright 2018-2022 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package drafts

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libcnb"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-github/v43/github"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/pipeline-builder/actions"
	"github.com/pkg/errors"
)

type Payload struct {
	PrimaryBuildpack Buildpack
	Builder          Builder
	NestedBuildpacks []Buildpack
	Release          Release
}

type Buildpack struct {
	libcnb.Buildpack
	OrderGroups  []libcnb.BuildpackOrder `toml:"order"`
	Dependencies []libpak.BuildpackDependency
}

type Builder struct {
	Description string
	Buildpacks  []struct {
		URI string
	}
	OrderGroups []libcnb.BuildpackOrder `toml:"order"`
	Stack       BuilderStack            `toml:"stack"`
}

type BuilderStack struct {
	ID         string `toml:"id"`
	BuildImage string `toml:"build-image"`
	RunImage   string `toml:"run-image"`
}

func (b Builder) Flatten() []string {
	tmp := []string{}

	for _, bp := range b.Buildpacks {
		tmp = append(tmp, strings.TrimPrefix(bp.URI, "docker://"))
	}

	return tmp
}

type Package struct {
	Dependencies []struct {
		URI string `toml:"uri"`
	}
}

func (b Package) Flatten() []string {
	tmp := []string{}

	for _, bp := range b.Dependencies {
		tmp = append(tmp, strings.TrimPrefix(bp.URI, "docker://"))
	}

	return tmp
}

type Release struct {
	ID   string
	Name string
	Body string
	Tag  string
}

//go:generate mockery --name BuildpackLoader --case=underscore

type BuildpackLoader interface {
	LoadBuildpack(id string) (Buildpack, error)
	LoadBuildpacks(uris []string) ([]Buildpack, error)
}

type Drafter struct {
	Loader BuildpackLoader
}

func (d Drafter) BuildAndWriteReleaseToFileDraftFromTemplate(outputPath, templateContents string, context interface{}) error {
	fp, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("unable to create file %s\n%w", outputPath, err)
	}
	defer fp.Close()

	return d.BuildAndWriteReleaseDraftFromTemplate(fp, templateContents, context)
}

func (d Drafter) BuildAndWriteReleaseDraftFromTemplate(output io.Writer, templateContents string, context interface{}) error {
	tmpl, err := template.New("draft").Parse(templateContents)
	if err != nil {
		return fmt.Errorf("unable to parse template %q\n%w", templateContents, err)
	}

	err = tmpl.Execute(output, context)
	if err != nil {
		return fmt.Errorf("unable to execute template %q\n%w", templateContents, err)
	}

	return nil
}

func (d Drafter) CreatePayload(inputs actions.Inputs, buildpackPath string) (Payload, error) {
	release := Release{
		ID:   inputs["release_id"],
		Name: inputs["release_name"],
		Body: inputs["release_body"],
		Tag:  inputs["release_tag_name"],
	}

	builder, err := loadBuilderTOML(buildpackPath)
	if err != nil {
		return Payload{}, err
	}

	if builder != nil {
		bps, err := d.Loader.LoadBuildpacks(builder.Flatten())
		if err != nil {
			return Payload{}, fmt.Errorf("unable to load buildpacks\n%w", err)
		}

		return Payload{
			PrimaryBuildpack: Buildpack{},
			Builder:          *builder,
			NestedBuildpacks: bps,
			Release:          release,
		}, nil
	}

	bp, err := loadBuildpackTOMLFromFile(buildpackPath)
	if err != nil {
		return Payload{}, err
	}

	pkg, err := loadPackage(buildpackPath)
	if err != nil {
		return Payload{}, err
	}

	if bp != nil && pkg == nil { // component
		return Payload{
			PrimaryBuildpack: *bp,
			Release:          release,
		}, nil
	} else if bp != nil && pkg != nil { // composite
		bps, err := d.Loader.LoadBuildpacks(pkg.Flatten())
		if err != nil {
			return Payload{}, fmt.Errorf("unable to load buildpacks\n%w", err)
		}

		return Payload{
			NestedBuildpacks: bps,
			PrimaryBuildpack: *bp,
			Release:          release,
		}, nil
	}

	return Payload{}, fmt.Errorf("unable to generate payload, need buildpack.toml or buildpack.toml + package.toml or builder.toml")
}

func loadBuildpackTOMLFromFile(buildpackPath string) (*Buildpack, error) {
	rawTOML, err := ioutil.ReadFile(filepath.Join(buildpackPath, "buildpack.toml"))
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to read buildpack toml\n%w", err)
	}

	return loadBuildpackTOML(rawTOML)
}

func loadBuildpackTOML(TOML []byte) (*Buildpack, error) {
	bp := &Buildpack{}
	if err := toml.Unmarshal(TOML, bp); err != nil {
		return nil, fmt.Errorf("unable to parse buildpack TOML\n%w", err)
	}

	if deps, found := bp.Metadata["dependencies"]; found {
		if depList, ok := deps.([]map[string]interface{}); ok {
			for _, dep := range depList {
				bpDep := libpak.BuildpackDependency{
					ID:      asString(dep, "id"),
					Name:    asString(dep, "name"),
					Version: asString(dep, "version"),
					URI:     asString(dep, "uri"),
					SHA256:  asString(dep, "sha256"),
					PURL:    asString(dep, "purl"),
				}

				if stacks, ok := dep["stacks"].([]interface{}); ok {
					for _, stack := range stacks {
						if stack, ok := stack.(string); ok {
							bpDep.Stacks = append(bpDep.Stacks, stack)
						}
					}
				}

				if cpes, ok := dep["cpes"].([]interface{}); ok {
					for _, cpe := range cpes {
						if cpe, ok := cpe.(string); ok {
							bpDep.CPEs = append(bpDep.CPEs, cpe)
						}
					}
				}

				if licenses, ok := dep["licenses"].([]map[string]interface{}); ok {
					for _, license := range licenses {
						bpDep.Licenses = append(bpDep.Licenses, libpak.BuildpackDependencyLicense{
							Type: asString(license, "type"),
							URI:  asString(license, "uri"),
						})
					}
				}

				bp.Dependencies = append(bp.Dependencies, bpDep)
			}
		} else {
			return nil, fmt.Errorf("unable to read dependencies from %v", bp.Metadata)
		}
	}

	return bp, nil
}

func asString(m map[string]interface{}, key string) string {
	if tmp, ok := m[key].(string); ok {
		return tmp
	}

	return ""
}

func loadPackage(buildpackPath string) (*Package, error) {
	rawTOML, err := ioutil.ReadFile(filepath.Join(buildpackPath, "package.toml"))
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to read package toml\n%w", err)
	}

	pkg := &Package{}
	if err := toml.Unmarshal(rawTOML, pkg); err != nil {
		return nil, fmt.Errorf("unable to parse package TOML\n%w", err)
	}

	return pkg, nil
}

func loadBuilderTOML(buildpackPath string) (*Builder, error) {
	rawTOML, err := ioutil.ReadFile(filepath.Join(buildpackPath, "builder.toml"))
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to read builder toml\n%w", err)
	}

	builder := &Builder{}
	if err := toml.Unmarshal(rawTOML, builder); err != nil {
		return nil, fmt.Errorf("unable to parse builder TOML\n%w", err)
	}

	return builder, nil
}

type GithubBuildpackLoader struct {
	GithubClient *github.Client
	RegexMappers []string
}

func (g GithubBuildpackLoader) LoadBuildpacks(uris []string) ([]Buildpack, error) {
	buildpacks := []Buildpack{}

	for _, uri := range uris {
		bp, err := g.LoadBuildpack(uri)
		if err != nil {
			return []Buildpack{}, fmt.Errorf("unable to process %s\n%w", uri, err)
		}
		buildpacks = append(buildpacks, bp)
	}

	return buildpacks, nil
}

func (g GithubBuildpackLoader) LoadBuildpack(imgUri string) (Buildpack, error) {
	uris, err := g.mapURIs(imgUri)
	if err != nil {
		return Buildpack{}, fmt.Errorf("unable to map URIs\n%w", err)
	}

	var tomlBytes []byte
	for _, uri := range uris {
		uriPattern := regexp.MustCompile(`.*\/(.*)\/(.*):(.*)`)

		parts := uriPattern.FindStringSubmatch(uri)
		if len(parts) != 4 {
			return Buildpack{}, fmt.Errorf("unable to parse %s, found %q", uri, parts)
		}

		org := parts[1]
		repo := parts[2]
		version := parts[3]
		if regexp.MustCompile(`\d+\.\d+\.\d+`).MatchString(version) {
			version = fmt.Sprintf("v%s", version)
		}

		tomlBytes, err = g.fetchBuildpackTOML(org, repo, version)
		if err != nil {
			var apiErr *github.ErrorResponse
			if errors.As(err, &apiErr) && apiErr.Response.StatusCode == 404 {
				fmt.Println("skipping 404", apiErr)
				continue
			}
			return Buildpack{}, fmt.Errorf("unable to fetch toml\n%w", err)
		}

		if len(tomlBytes) > 0 {
			break
		}
	}

	if len(tomlBytes) == 0 {
		return Buildpack{}, fmt.Errorf("unable to fetch toml, file not found")
	}

	bp, err := loadBuildpackTOML(tomlBytes)
	if err != nil {
		return Buildpack{}, fmt.Errorf("unable to load buildpack toml from image\n%w", err)
	}

	return *bp, nil
}

func (g GithubBuildpackLoader) mapURIs(uri string) ([]string, error) {
	possibilities := []string{uri}

	for _, mapper := range g.RegexMappers {
		if len(mapper) <= 3 {
			continue
		}

		splitCh := string(mapper[0])
		parts := strings.SplitN(mapper[1:len(mapper)-1], splitCh, 2)

		expr, err := regexp.Compile(parts[0])
		if err != nil {
			return []string{}, fmt.Errorf("unable to parse regex %s\n%w", mapper, err)
		}

		possibilities = append(possibilities, expr.ReplaceAllString(uri, parts[1]))
	}

	return possibilities, nil
}

func (g GithubBuildpackLoader) fetchBuildpackTOML(org, repo, version string) ([]byte, error) {
	body, _, err := g.GithubClient.Repositories.DownloadContents(
		context.Background(),
		org,
		repo,
		"/buildpack.toml",
		&github.RepositoryContentGetOptions{Ref: version})
	if err != nil {
		return []byte{}, fmt.Errorf("unable to download file\n%w", err)
	}

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, body)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to read downloaded file\n%w", err)
	}

	return buf.Bytes(), nil
}

type RegistryBuildpackLoader struct{}

func (r RegistryBuildpackLoader) LoadBuildpacks(uris []string) ([]Buildpack, error) {
	buildpacks := []Buildpack{}

	for _, uri := range uris {
		bp, err := r.LoadBuildpack(uri)
		if err != nil {
			return []Buildpack{}, fmt.Errorf("unable to process %s\n%w", uri, err)
		}
		buildpacks = append(buildpacks, bp)
	}

	return buildpacks, nil
}

func (r RegistryBuildpackLoader) LoadBuildpack(uri string) (Buildpack, error) {
	if err := os.MkdirAll("/tmp", 1777); err != nil {
		return Buildpack{}, fmt.Errorf("unable to create /tmp\n%w", err)
	}

	tarFile, err := ioutil.TempFile("/tmp", "tarfiles")
	if err != nil {
		return Buildpack{}, fmt.Errorf("unable to create tempfile\n%w", err)
	}
	defer os.Remove(tarFile.Name())

	err = loadBuildpackImage(uri, tarFile)
	if err != nil {
		return Buildpack{}, fmt.Errorf("unable to load %s\n%w", uri, err)
	}

	_, err = tarFile.Seek(0, 0)
	if err != nil {
		return Buildpack{}, fmt.Errorf("unable to reset file pointer\n%w", err)
	}

	bpTOML, err := readBuildpackTOML(tarFile)
	if err != nil {
		return Buildpack{}, err
	}

	bp, err := loadBuildpackTOML(bpTOML)
	if err != nil {
		return Buildpack{}, fmt.Errorf("unable to load buildpack toml from image\n%w", err)
	}

	return *bp, nil
}

func loadBuildpackImage(ref string, to io.Writer) error {
	reference, err := name.ParseReference(ref)
	if err != nil {
		return fmt.Errorf("unable to parse reference for existing buildpack tag\n%w", err)
	}

	authn := authn.Anonymous
	if gcrKey, found := os.LookupEnv("JAVA_GCLOUD_SERVICE_ACCOUNT_KEY"); found {
		authn = google.NewJSONKeyAuthenticator(gcrKey)
	}

	img, err := remote.Image(reference, remote.WithAuth(authn))
	if err != nil {
		return fmt.Errorf("unable to fetch remote image\n%w", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return fmt.Errorf("unable to fetch layer\n%w", err)
	}

	if len(layers) == 1 {
		l := layers[0]
		rc, err := l.Uncompressed()
		if err != nil {
			return fmt.Errorf("unable to get uncompressed reader\n%w", err)
		}
		_, err = io.Copy(to, rc)
		return err
	}

	fs := mutate.Extract(img)
	_, err = io.Copy(to, fs)
	return err
}

func readBuildpackTOML(tarFile *os.File) ([]byte, error) {
	t := tar.NewReader(tarFile)
	for {
		f, err := t.Next()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return []byte{}, fmt.Errorf("unable to read TAR file\n%w", err)
		}

		if strings.HasSuffix(f.Name, "buildpack.toml") {
			info := f.FileInfo()

			if info.IsDir() || (info.Mode()&os.ModeSymlink != 0) {
				return []byte{}, fmt.Errorf("unable to read buildpack.toml, unexpected file type (directory or symlink)")
			}

			buf := &bytes.Buffer{}
			_, err := io.Copy(buf, t)
			if err != nil {
				return []byte{}, fmt.Errorf("unable to read buildpack.toml\n%w", err)
			}

			return buf.Bytes(), nil
		}
	}

	return []byte{}, fmt.Errorf("unable to find buildpack.toml in image")
}
