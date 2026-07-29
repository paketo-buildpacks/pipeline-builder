package octo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/disiqueira/gotree"
)

type BuildScriptData struct {
	Helpers       map[string]string
	Architectures []string
}

func ContributeScripts(descriptor Descriptor) ([]Contribution, error) {
	scriptPath := filepath.Join(descriptor.Path, "scripts", "build.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, nil
	}

	c := Contribution{
		Path:        "scripts/build.sh",
		Permissions: 0755,
	}

	arches, err := getBuildArchitectures(filepath.Join(descriptor.Path, "buildpack.toml"))
	if err != nil {
		return nil, err
	}

	templateContents := []byte(StatikString("/build-script.sh"))
	tmpl, err := template.New("build.sh").Parse(string(templateContents))
	if err != nil {
		return nil, fmt.Errorf("unable to parse template %q\n%w", templateContents, err)
	}

	output := &bytes.Buffer{}
	err = tmpl.Execute(output, BuildScriptData{
		Helpers:       descriptor.Helpers,
		Architectures: arches,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute template %q\n%w", templateContents, err)
	}

	c.Content = output.Bytes()
	c.Structure = gotree.New("scripts/ [project scripts]")
	c.Structure.Add("scripts/build.sh [build]")

	return []Contribution{c}, nil
}

func getBuildArchitectures(buildpackTOMLPath string) ([]string, error) {
	type target struct {
		Arch string `toml:"arch"`
		OS   string `toml:"os"`
	}
	var bp struct {
		Targets []target `toml:"targets"`
	}

	b, err := os.ReadFile(buildpackTOMLPath)
	if os.IsNotExist(err) {
		return []string{"amd64", "arm64"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to read %s\n%w", buildpackTOMLPath, err)
	}

	if err := toml.Unmarshal(b, &bp); err != nil {
		return nil, fmt.Errorf("unable to decode %s\n%w", buildpackTOMLPath, err)
	}

	if len(bp.Targets) == 0 {
		return []string{"amd64", "arm64"}, nil
	}

	seen := make(map[string]bool)
	var arches []string
	for _, t := range bp.Targets {
		if !seen[t.Arch] {
			arches = append(arches, t.Arch)
			seen[t.Arch] = true
		}
	}

	sort.Strings(arches)
	return arches, nil
}
