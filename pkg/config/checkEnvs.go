package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func CheckEnvs() error {
	// Override configuration if env vars are defined
	if os.Getenv("HCLOUD_TOKEN") != "" {
		viper.Set("HCLOUD_TOKEN", os.Getenv("HCLOUD_TOKEN"))
	}
	if os.Getenv("SERVER_ID") != "" {
		viper.Set("SERVER_ID", os.Getenv("SERVER_ID"))
	}
	if os.Getenv("TOP_SERVER_NAME") != "" {
		viper.Set("TOP_SERVER_NAME", os.Getenv("TOP_SERVER_NAME"))
	}
	if os.Getenv("BASE_SERVER_NAME") != "" {
		viper.Set("BASE_SERVER_NAME", os.Getenv("BASE_SERVER_NAME"))
	}
	if os.Getenv("HOUR_START") != "" {
		viper.Set("HOUR_START", os.Getenv("HOUR_START"))
	}
	if os.Getenv("HOUR_STOP") != "" {
		viper.Set("HOUR_STOP", os.Getenv("HOUR_STOP"))
	}

	if viper.GetString("HCLOUD_TOKEN") == "" ||
		viper.GetInt("SERVER_ID") == 0 ||
		viper.GetString("TOP_SERVER_NAME") == "" ||
		viper.GetString("BASE_SERVER_NAME") == "" ||
		viper.GetString("HOUR_START") == "" ||
		viper.GetString("HOUR_STOP") == "" {
		return fmt.Errorf("missing or incomplete configuration")
	}

	return nil
}
