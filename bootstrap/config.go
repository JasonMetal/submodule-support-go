// Package config 负责配置信息
package bootstrap

import (
	"fmt"
	viperlib "github.com/spf13/viper" // 读取配置文件的第三方开源库,自定义包名，避免与内置 viper 实例冲突
)

type Config struct {
}

// viper 库实例
var viper *viperlib.Viper

// ConfigFunc 动态加载配置信息
type ConfigFunc func() map[string]interface{}

// ConfigFuncs 先加载到此数组，loadConfig 在动态生成配置信息
var ConfigFuncs map[string]ConfigFunc

func init() {
	viper = viperlib.New()
	viper.SetConfigType("yaml")
}

func GetConfig(env string) *viperlib.Viper {
	path := fmt.Sprintf("./config/%s/", env)
	fmt.Println(path)
	viper.AddConfigPath(path)    // 设置文件所在路径
	viper.SetConfigName("mysql") // 设置文件名称（无后缀）

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viperlib.ConfigFileNotFoundError); ok {
			panic(" Config file not found; ignore error if desired")
		} else {
			panic("Config file was found but another error was produced")
		}
	}
	viper.WatchConfig()
	return viper
}
