package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var keys = []string{
	"HCLOUD_TOKEN",
	"SERVER_ID",
	"TOP_SERVER_NAME",
	"BASE_SERVER_NAME",
	"HOUR_START",
	"HOUR_STOP",
	"WEEK_DAYS",
}

func CheckEnvs() error {
	// Override configuration if env vars are defined
	for _, key := range keys {
		if os.Getenv(key) != "" {
			viper.Set(key, os.Getenv(key))
		}

		if viper.GetString(key) == "" {
			return fmt.Errorf("%s is not defined", key)
		}
	}

	return nil
}
