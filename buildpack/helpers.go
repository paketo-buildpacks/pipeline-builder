package buildpack

import (
	"encoding/json"
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
)

func GetStringLabel(image v1.Image, key string) (string, error) {
	configFile, err := configFile(image)
	if err != nil {
		return "", fmt.Errorf("unable to load config for get\n%w", err)
	}

	config := configFile.Config.DeepCopy()

	stringValue, ok := config.Labels[key]
	if !ok {
		return "", fmt.Errorf("unable to find label %s", key)
	}

	return stringValue, nil
}

func GetLabel(image v1.Image, key string, value interface{}) error {
	stringValue, err := GetStringLabel(image, key)
	if err != nil {
		return fmt.Errorf("unable to get string label %s\n%w", key, err)
	}
	return json.Unmarshal([]byte(stringValue), value)
}

func SetLabels(image v1.Image, labels map[string]interface{}) (v1.Image, error) {
	configFile, err := image.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("unable to load config for set\n%s", err)
	}

	config := *configFile.Config.DeepCopy()
	if config.Labels == nil {
		config.Labels = map[string]string{}
	}

	for k, v := range labels {
		dataBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal data to JSON for label %s", k)
		}

		config.Labels[k] = string(dataBytes)
	}

	return mutate.Config(image, config)
}

func configFile(image v1.Image) (*v1.ConfigFile, error) {
	cfg, err := image.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("unable to get image config")
	} else if cfg == nil {
		return nil, fmt.Errorf("unable to get non-nil image config")
	}
	return cfg, nil
}
