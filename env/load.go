package env

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/spf13/viper"
)

var config = defaultConfig()

func LoadConfig(configFile string) {
	viper.AutomaticEnv()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	content, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}

	err = viper.MergeConfig(bytes.NewReader(content))
	if err != nil {
		panic(err)
	}

	if configFile != "" {
		content, err = os.ReadFile(configFile)
		if err != nil {
			panic(err)
		}

		err = viper.MergeConfig(bytes.NewReader(content))
		if err != nil {
			panic(err)
		}
	}

	if err := viper.Unmarshal(config); err != nil {
		panic(err)
	}
}
