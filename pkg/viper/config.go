package viper

import (
	"github.com/chroblert/jlog"
	"github.com/spf13/viper"
)

func YamlConfig() (scfurl, proxyport string) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetConfigPermissions(0644)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.Set("scfurl", "")
			viper.Set("proxyport", "1080")
			err := viper.SafeWriteConfig()
			if err != nil {
				jlog.Fatal(err)
			}
		} else {
			jlog.Println("read config error")
		}
		jlog.Fatal(err)
	}
	scfurl = viper.GetString("scfurl")
	proxyport = viper.GetString("proxyport")
	if scfurl == "" {
		jlog.Fatal("云函数地址为空")
	}
	if proxyport == "" {
		jlog.Fatal("本地代理端口为空")
	}
	return scfurl, proxyport
}
