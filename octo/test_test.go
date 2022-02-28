package octo_test

import (
	"github.com/paketo-buildpacks/pipeline-builder/octo"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// use our own type as the events cannot be unmarshalled
type jobs struct {
	Jobs map[string]*actions.Job `yaml:jobs`
}

func TestContributeTest_Default(t *testing.T) {
	dir, descriptor := setUp(t)
	defer os.RemoveAll(dir)

	err := ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte{}, 0644)
	assert.NoError(t, err)

	contribution, err := octo.ContributeTest(descriptor)
	assert.NoError(t, err)

	assert.Equal(t, ".github/workflows/tests.yml", contribution.Path)
	t.Log(string(contribution.Content))

	var workflow jobs
	err = yaml.Unmarshal(contribution.Content, &workflow)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(workflow.Jobs))
	assert.NotNil(t, workflow.Jobs["unit"])
	assert.Nil(t, workflow.Jobs["integration"])

	steps := workflow.Jobs["unit"].Steps
	assert.Contains(t, steps[len(steps) - 1].Run, "richgo test ./...")
}


func TestContributeTest_IntegrationTests(t *testing.T) {
	dir, descriptor := setUp(t)
	defer os.RemoveAll(dir)

	err := ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte{}, 0644)
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(dir, "integration"), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(dir, "integration", "main.go"), []byte{}, 0644)
	assert.NoError(t, err)

	contribution, err := octo.ContributeTest(descriptor)
	assert.NoError(t, err)

	assert.Equal(t, ".github/workflows/tests.yml", contribution.Path)
	t.Log(string(contribution.Content))

	var workflow jobs
	err = yaml.Unmarshal(contribution.Content, &workflow)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(workflow.Jobs))
	assert.NotNil(t, workflow.Jobs["unit"])
	assert.NotNil(t, workflow.Jobs["integration"])

	steps := workflow.Jobs["unit"].Steps
	assert.Contains(t, steps[len(steps) - 1].Run, "richgo test -short ./...")

	steps = workflow.Jobs["integration"].Steps
	assert.Contains(t, steps[len(steps) - 1].Run, "richgo test ./integration/...")
}

func setUp(t *testing.T) (string, octo.Descriptor ){
	dir, err := ioutil.TempDir("", "main-package")
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(dir, ".github"), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(dir, ".github", "pipeline-descriptor.yaml"), []byte(`---
github:
  username: ${{ secrets.JAVA_GITHUB_USERNAME }}
  token:    ${{ secrets.JAVA_GITHUB_TOKEN }}

package:
  repository:     gcr.io/paketo-buildpacks/dummy
  register:       true
  registry_token: ${{ secrets.JAVA_GITHUB_TOKEN }}
`), 0644)
	assert.NoError(t, err)

	descriptor, err := octo.NewDescriptor(filepath.Join(dir, ".github", "pipeline-descriptor.yaml"))
	assert.NoError(t, err)

	t.Logf("%+v", descriptor)


	assert.NotNil(t, descriptor.Package)

	return dir, descriptor
}
