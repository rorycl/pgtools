package main

import (
	"fmt"

	yaml "gopkg.in/yaml.v3"
)

// Config represents the configurations set out in a yaml file
type Config map[string]DBQueryGroupConfig

// DBQueryGroupConfig sets out the configuration items for each group of
// databases
type DBQueryGroupConfig struct {
	Databases   []string
	Concurrency int
	Iterations  int
	Queries     []string
}

// LoadYaml loads a yaml file and returns a Settings structure
func LoadYaml(yamlByte []byte) (Config, error) {
	var config Config
	err := yaml.Unmarshal(yamlByte, &config)
	if err != nil {
		return config, err
	}

	err = config.check()
	return config, err
}

// check checks the validity of the settings file
func (c Config) check() error {
	for k, v := range c {
		if v.Concurrency == 0 {
			return fmt.Errorf("group %s requires more than 0 concurrency", k)
		}
		if v.Concurrency > len(v.Databases) {
			return fmt.Errorf(
				"group %s has concurrency %d greater than length of databases %d",
				k, v.Concurrency, len(v.Databases))
		}
		if v.Iterations == 0 {
			return fmt.Errorf("group %s requires 1 or more iterations", k)
		}
		if len(v.Queries) == 0 {
			return fmt.Errorf("group %s has no queries defined", k)
		}
	}
	return nil
}
