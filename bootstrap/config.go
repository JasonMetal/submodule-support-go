// Package config 负责配置信息
package bootstrap

import (
	"fmt"
	"idea-go/helpers/config"
)

type Config struct {
}

func getDbNames(filename string) []string {
	DbNames := make([]string, 0)
	path := fmt.Sprintf("./config/%s/%s.yml", DevEnv, filename)

	DBConfigs, err := config.GetConfig(path)

	configList, err := DBConfigs.Map(filename)
	if err == nil {
		for DBName, _ := range configList {
			DbNames = append(DbNames, DBName)
		}
	}

	return DbNames
}
