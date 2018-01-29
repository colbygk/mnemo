package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"okea/services/models"
	"os"
)

func ReadNGINXConfig(yaml_file_name string) (*Nginx_yaml, error) {
	var new_config Nginx_yaml

	yaml_bytes, err := ioutil.ReadFile(yaml_file_name)
	if err != nil {
		log.Fatalf(" ReadFile: %s %s\n", yaml_file_name, err)
		return nil, err
	}

	err = yaml.Unmarshal(yaml_bytes, &new_config)
	if err != nil {
		log.Fatalf(" Parsing YAML %s: %v\n", yaml_file_name, err)
		return nil, err
	}

	return &new_config, nil
}

func WriteNGINXConfig(yaml_file_name string, config *Nginx_yaml) error {

	yaml_bytes, err := yaml.Marshal(config)

	err = ioutil.WriteFile(yaml_file_name, yaml_bytes, os.FileMode(0644))
	if err != nil {
		log.Fatalf(" Writing YAML %s: %v\n", yaml_file_name, err)
	}

	return err
}
