package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

/*
type Config struct {
	Firewall_network_rules map[string]Options
}

type Options struct {
	Src string
	Dst string
}

*/
type NGINX_vhosts struct {
	Listen          string
	ServerName      string `yaml:"server_name"`
	Root            string
	Index           string
	ErrorPage       string `yaml:"error_page"`
	AccessLog       string `yaml:"access_log"`
	ErrorLog        string `yaml:"error_log"`
	ExtraParameters string `yaml:"extra_parameters"`
}

type NGINX_yaml struct {
	VHosts []NGINX_vhosts `yaml:"nginx_vhosts"`
}

func main() {
	filename, _ := filepath.Abs("./nginx.yml")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var config NGINX_yaml

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Value: %#v\n", config.VHosts)
}
