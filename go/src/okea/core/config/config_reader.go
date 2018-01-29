package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type JSONConfig struct {
	Hostname          string `json:"hostname"`
	Port              int    `json:"port"`
	Okea_username     string `json:"okea_username"`
	TrustTLS          bool   `json:"trusttls"`
	Infoblox_username string `json:"infoblox_username"`
	Infoblox_password string `json:"infoblox_password"`
	Custom_Root_CA    string `json:"custom_root_ca"`
	TLS_cert_filename string `json:"tls_cert_filename"`
	TLS_key_filename  string `json:"tls_key_filename"`
}

func LoadConfig(conf_filename string) (JSONConfig, error) {

	file, err := ioutil.ReadFile(conf_filename)
	var newconfig JSONConfig

	if err != nil {
		// log.Printf("Config file read error: %v\n", err)
		return newconfig, err
	} else {

		err = json.Unmarshal(file, &newconfig)

		if err != nil {
			// log.Printf("Unknown JSON: %v\n", err)
		}
	}

	return newconfig, err
}

func (jc *JSONConfig) WriteConfig(conf_filename string) error {
	b, err := json.MarshalIndent(jc, "", "  ")
	if err != nil {
		log.Printf("Unknown JSON: %v\n", err)
	}
	return ioutil.WriteFile(conf_filename, b, 0600)
}
