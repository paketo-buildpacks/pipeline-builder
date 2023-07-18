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
	"sort"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libcnb"
	"github.com/google/go-github/v43/github"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/pipeline-builder/actions"
	"github.com/pkg/errors"
)

type Payload struct {
	PrimaryBuildpack Buildpack
	PrimaryExtension Extension
	Builder          Builder
	NestedBuildpacks []Buildpack
	NestedExtensions []Extension
	Release          Release
}

type Buildpack struct {
	libcnb.Buildpack
	OrderGroups  []libcnb.BuildpackOrder `toml:"order"`
	Dependencies []libpak.BuildModuleDependency
}

type Extension struct {
	libcnb.Extension
	OrderGroups  []libcnb.BuildpackOrder `toml:"order-extensions"`
	Dependencies []libpak.BuildModuleDependency
}

type Builder struct {
	Description string
	Extensions  []struct {
		URI string
	}
	Buildpacks []struct {
		URI string
	}
	ExtensionsOrderGroups []libcnb.BuildpackOrder `toml:"order-extensions"`
	OrderGroups           []libcnb.BuildpackOrder `toml:"order"`
	Stack                 BuilderStack            `toml:"stack"`
}

type BuilderStack struct {
	ID         string `toml:"id"`
	BuildImage string `toml:"build-image"`
	RunImage   string `toml:"run-image"`
}

func (b Builder) FlattenBuildpacks() []string {
	tmp := []string{}

	for _, bp := range b.Buildpacks {
		tmp = append(tmp, strings.TrimPrefix(bp.URI, "docker://"))
	}

	return tmp
}

func (b Builder) FlattenExtensions() []string {
	tmp := []string{}

	for _, xtn := range b.Extensions {
		tmp = append(tmp, strings.TrimPrefix(xtn.URI, "docker://"))
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

//go:generate mockery --name BuildModuleLoader --case=underscore

type BuildModuleLoader interface {
	LoadBuildpack(id string) (Buildpack, error)
	LoadBuildpacks(uris []string) ([]Buildpack, error)
	LoadExtension(id string) (Extension, error)
	LoadExtensions(uris []string) ([]Extension, error)
}

type Drafter struct {
	Loader BuildModuleLoader
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

func (d Drafter) CreatePayload(inputs actions.Inputs, buildModulePath string) (Payload, error) {
	release := Release{
		ID:   inputs["release_id"],
		Name: inputs["release_name"],
		Body: inputs["release_body"],
		Tag:  inputs["release_tag_name"],
	}

	builder, err := loadBuilderTOML(buildModulePath)
	if err != nil {
		return Payload{}, err
	}

	// If we have a builder.toml the Payload is for a Builder
	if builder != nil {
		bps, err := d.Loader.LoadBuildpacks(builder.FlattenBuildpacks())
		if err != nil {
			return Payload{}, fmt.Errorf("unable to load buildpacks\n%w", err)
		}

		xtns, err := d.Loader.LoadExtensions(builder.FlattenExtensions())
		if err != nil {
			return Payload{}, fmt.Errorf("unable to load extensions\n%w", err)
		}

		return Payload{
			PrimaryBuildpack: Buildpack{},
			Builder:          *builder,
			NestedBuildpacks: bps,
			NestedExtensions: xtns,
			Release:          release,
		}, nil
	}

	// If we are not processing a Builder, then we can be
	// an Extension, or a Composite Buildpack, or a Single Buildpack

	bp, err := loadBuildpackTOMLFromFile(buildModulePath)
	if err != nil {
		return Payload{}, err
	}

	xtn, err := loadExtensionTOMLFromFile(buildModulePath)
	if err != nil {
		return Payload{}, err
	}

	pkg, err := loadPackage(buildModulePath)
	if err != nil {
		return Payload{}, err
	}

	if bp != nil && pkg == nil { // single buildpack
		return Payload{
			PrimaryBuildpack: *bp,
			Release:          release,
		}, nil
	} else if xtn != nil && pkg == nil { // single extension
		return Payload{
			PrimaryExtension: *xtn,
			Release:          release,
		}, nil
	} else if bp != nil && pkg != nil { // composite buildpack
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

func loadBuildpackTOMLFromFile(buildModulePath string) (*Buildpack, error) {
	rawTOML, err := ioutil.ReadFile(filepath.Join(buildModulePath, "buildpack.toml"))
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to read buildpack toml\n%w", err)
	}

	return loadBuildpackTOML(rawTOML)
}

func loadExtensionTOMLFromFile(buildModulePath string) (*Extension, error) {
	rawTOML, err := ioutil.ReadFile(filepath.Join(buildModulePath, "extension.toml"))
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to read extension toml\n%w", err)
	}

	return loadExtensionTOML(rawTOML)
}

func processDependencies(depList []map[string]interface{}, isExtension bool) []libpak.BuildModuleDependency {
	deps := []libpak.BuildModuleDependency{}
	for _, dep := range depList {
		bpDep := libpak.BuildModuleDependency{
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
				bpDep.Licenses = append(bpDep.Licenses, libpak.BuildModuleDependencyLicense{
					Type: asString(license, "type"),
					URI:  asString(license, "uri"),
				})
			}
		}

		deps = append(deps, bpDep)
	}
	return deps
}

func loadBuildpackTOML(TOML []byte) (*Buildpack, error) {
	bp := &Buildpack{}
	if err := toml.Unmarshal(TOML, bp); err != nil {
		return nil, fmt.Errorf("unable to parse buildpack TOML\n%w", err)
	}
	sort.Slice(bp.Stacks, func(i, j int) bool {
		return strings.ToLower(bp.Stacks[i].ID) < strings.ToLower(bp.Stacks[j].ID)
	})

	if deps, found := bp.Metadata["dependencies"]; found {
		if depList, ok := deps.([]map[string]interface{}); ok {
			bp.Dependencies = processDependencies(depList, false)
		} else {
			return nil, fmt.Errorf("unable to read dependencies from %v", bp.Metadata)
		}

		sort.Slice(bp.Dependencies, func(i, j int) bool {
			return strings.ToLower(bp.Dependencies[i].Name) < strings.ToLower(bp.Dependencies[j].Name)
		})
	}

	return bp, nil
}

func loadExtensionTOML(TOML []byte) (*Extension, error) {
	xtn := &Extension{}
	if err := toml.Unmarshal(TOML, xtn); err != nil {
		return nil, fmt.Errorf("unable to parse extension TOML\n%w", err)
	}

	if deps, found := xtn.Metadata["dependencies"]; found {
		if depList, ok := deps.([]map[string]interface{}); ok {
			xtn.Dependencies = processDependencies(depList, true)
		} else {
			return nil, fmt.Errorf("unable to read dependencies from %v", xtn.Metadata)
		}

		sort.Slice(xtn.Dependencies, func(i, j int) bool {
			return strings.ToLower(xtn.Dependencies[i].Name) < strings.ToLower(xtn.Dependencies[j].Name)
		})
	}

	return xtn, nil
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

func loadBuilderTOML(builderPath string) (*Builder, error) {
	rawTOML, err := ioutil.ReadFile(filepath.Join(builderPath, "builder.toml"))
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

type GithubBuildModuleLoader struct {
	GithubClient *github.Client
	RegexMappers []string
}

func (g GithubBuildModuleLoader) LoadBuildpacks(uris []string) ([]Buildpack, error) {
	buildpacks := []Buildpack{}

	for _, uri := range uris {
		bp, err := g.LoadBuildpack(uri)
		if err != nil {
			return []Buildpack{}, fmt.Errorf("unable to process %s\n%w", uri, err)
		}
		buildpacks = append(buildpacks, bp)
	}

	sort.Slice(buildpacks, func(i, j int) bool {
		return strings.ToLower(buildpacks[i].Info.Name) < strings.ToLower(buildpacks[j].Info.Name)
	})

	return buildpacks, nil
}

func (g GithubBuildModuleLoader) LoadExtensions(uris []string) ([]Extension, error) {
	extensions := []Extension{}

	for _, uri := range uris {
		xtn, err := g.LoadExtension(uri)
		if err != nil {
			return []Extension{}, fmt.Errorf("unable to process %s\n%w", uri, err)
		}
		extensions = append(extensions, xtn)
	}

	sort.Slice(extensions, func(i, j int) bool {
		return strings.ToLower(extensions[i].Info.Name) < strings.ToLower(extensions[j].Info.Name)
	})

	return extensions, nil
}

func (g GithubBuildModuleLoader) LoadBuildpack(imgUri string) (Buildpack, error) {
	uris, err := g.mapURIs(imgUri)
	if err != nil {
		return Buildpack{}, fmt.Errorf("unable to map URIs\n%w", err)
	}

	origOrg, origRepo, _, err := parseRepoOrgVersionFromImageUri(imgUri)
	if err != nil {
		return Buildpack{}, fmt.Errorf("unable to parse original image uri\n%w", err)
	}

	for _, uri := range uris {
		org, repo, version, err := parseRepoOrgVersionFromImageUri(uri)
		if err != nil {
			return Buildpack{}, fmt.Errorf("unable to parse image uri\n%w", err)
		}

		paths, err := g.mapBuildpackTOMLPath(origOrg, origRepo)
		if err != nil {
			return Buildpack{}, fmt.Errorf("unable to map buildpack toml path\n%w", err)
		}

		if regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(version) {
			version = fmt.Sprintf("v%s", version)
		}

		for _, path := range paths {
			tomlBytes, err := g.fetchTOMLFile(org, repo, version, path)
			if err != nil {
				var apiErr *github.ErrorResponse
				if errors.As(err, &apiErr) && apiErr.Response.StatusCode == 404 {
					fmt.Println("skipping 404", apiErr)
					continue
				}
				return Buildpack{}, fmt.Errorf("unable to fetch toml\n%w", err)
			}

			if len(tomlBytes) > 0 {
				bp, err := loadBuildpackTOML(tomlBytes)
				if err != nil {
					return Buildpack{}, fmt.Errorf("unable to load buildpack toml from image\n%w", err)
				}

				bp.Info.Version = version
				return *bp, nil
			}
		}
	}

	return Buildpack{}, fmt.Errorf("unable to load buildpack.toml for %s", imgUri)
}

func (g GithubBuildModuleLoader) LoadExtension(imgUri string) (Extension, error) {
	uris, err := g.mapURIs(imgUri)
	if err != nil {
		return Extension{}, fmt.Errorf("unable to map URIs\n%w", err)
	}

	origOrg, origRepo, _, err := parseRepoOrgVersionFromImageUri(imgUri)
	if err != nil {
		return Extension{}, fmt.Errorf("unable to parse original image uri\n%w", err)
	}

	for _, uri := range uris {
		org, repo, version, err := parseRepoOrgVersionFromImageUri(uri)
		if err != nil {
			return Extension{}, fmt.Errorf("unable to parse image uri\n%w", err)
		}

		paths, err := g.mapExtensionTOMLPath(origOrg, origRepo)
		if err != nil {
			return Extension{}, fmt.Errorf("unable to map extension toml path\n%w", err)
		}

		if regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(version) {
			version = fmt.Sprintf("v%s", version)
		}

		for _, path := range paths {
			tomlBytes, err := g.fetchTOMLFile(org, repo, version, path)
			if err != nil {
				var apiErr *github.ErrorResponse
				if errors.As(err, &apiErr) && apiErr.Response.StatusCode == 404 {
					fmt.Println("skipping 404", apiErr)
					continue
				}
				return Extension{}, fmt.Errorf("unable to fetch toml\n%w", err)
			}

			if len(tomlBytes) > 0 {
				xtn, err := loadExtensionTOML(tomlBytes)
				if err != nil {
					return Extension{}, fmt.Errorf("unable to load buildpack toml from image\n%w", err)
				}

				xtn.Info.Version = version
				return *xtn, nil
			}
		}
	}

	return Extension{}, fmt.Errorf("unable to load buildpack.toml for %s", imgUri)
}

func (g GithubBuildModuleLoader) LoadPackages(imgUri string) (Package, error) {
	uris, err := g.mapURIs(imgUri)
	if err != nil {
		return Package{}, fmt.Errorf("unable to map URIs\n%w", err)
	}

	for _, uri := range uris {

		uriPattern := regexp.MustCompile(`.*\/(.*)\/(.*):(.*)`)

		parts := uriPattern.FindStringSubmatch(uri)
		if len(parts) != 4 {
			return Package{}, fmt.Errorf("unable to parse %s, found %q", uri, parts)
		}

		org := parts[1]
		repo := parts[2]
		version := parts[3]
		if regexp.MustCompile(`\d+\.\d+\.\d+`).MatchString(version) {
			version = fmt.Sprintf("v%s", version)
		}

		tomlBytes, err := g.fetchTOMLFile(org, repo, version, "/package.toml")
		if err != nil {
			var apiErr *github.ErrorResponse
			if errors.As(err, &apiErr) && apiErr.Response.StatusCode == 404 {
				fmt.Println("skipping 404", apiErr)
				continue
			}
			return Package{}, fmt.Errorf("unable to fetch toml\n%w", err)
		}

		if len(tomlBytes) > 0 {
			pkg := &Package{}
			if err := toml.Unmarshal(tomlBytes, pkg); err != nil {
				return Package{}, fmt.Errorf("unable to parse package TOML\n%w", err)
			}
			return *pkg, nil
		}
	}

	return Package{}, fmt.Errorf("unable to load package.toml for %s", imgUri)
}

func (g GithubBuildModuleLoader) mapURIs(uri string) ([]string, error) {
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

func (g GithubBuildModuleLoader) mapBuildpackTOMLPath(org, repo string) ([]string, error) {
	paths := []string{
		"/buildpack.toml",
	}

	org = strings.ToUpper(strings.ReplaceAll(org, "-", "_"))
	repo = strings.ToUpper(strings.ReplaceAll(repo, "-", "_"))

	if p, found := os.LookupEnv(fmt.Sprintf("BP_TOML_PATH_%s_%s", org, repo)); found {
		if !strings.HasSuffix(p, "/buildpack.toml") {
			p = fmt.Sprintf("%s/buildpack.toml", p)
		}
		return []string{p}, nil
	}

	return paths, nil
}

func (g GithubBuildModuleLoader) mapExtensionTOMLPath(org, repo string) ([]string, error) {
	paths := []string{
		"/extension.toml",
	}

	org = strings.ToUpper(strings.ReplaceAll(org, "-", "_"))
	repo = strings.ToUpper(strings.ReplaceAll(repo, "-", "_"))

	if p, found := os.LookupEnv(fmt.Sprintf("XTN_TOML_PATH_%s_%s", org, repo)); found {
		if !strings.HasSuffix(p, "/extension.toml") {
			p = fmt.Sprintf("%s/extension.toml", p)
		}
		return []string{p}, nil
	}

	return paths, nil
}

func (g GithubBuildModuleLoader) fetchTOMLFile(org, repo, version, path string) ([]byte, error) {
	fmt.Println("Fetching from org:", org, "repo:", repo, "version:", version, "path:", path)
	body, _, err := g.GithubClient.Repositories.DownloadContents(
		context.Background(),
		org,
		repo,
		path,
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

func parseRepoOrgVersionFromImageUri(imgUri string) (string, string, string, error) {
	uriPattern := regexp.MustCompile(`.*\/(.*)\/(.*):(.*)`)

	parts := uriPattern.FindStringSubmatch(imgUri)
	if len(parts) != 4 {
		return "", "", "", fmt.Errorf("unable to parse %s, found %q", imgUri, parts)
	}

	return parts[1], parts[2], parts[3], nil
}
