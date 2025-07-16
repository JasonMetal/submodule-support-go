// Package config 负责配置信息
package config

import (
	configLib "github.com/olebedev/config"
	"gopkg.in/yaml.v2"
	"os"
)

func GetConfig(filename string) (*configLib.Config, error) {
	var conf *configLib.Config

	conf, err := configLib.ParseYamlFile(filename)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func GetYaml(filePath string, configData any) any {
	//Deprecated
	//config, err := ioutil.ReadFile(filePath)
	config, err := os.ReadFile(filePath)
	if err != nil {
		return &configData
	}
	err = yaml.Unmarshal(config, &configData)
	if err != nil {
		return &configData
	}
	return configData
}
