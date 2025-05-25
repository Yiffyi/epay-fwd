package misc

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func SetupConfig() {
	// Set the file name and path (without extension)
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".") // Look for the configuration file in current directory.

	// Set defaults optionally
	viper.SetDefault("listen_addr", ":1323")
	viper.SetDefault("site_url", "")

	viper.SetDefault("alipay.app_id", "")
	viper.SetDefault("alipay.app_private_key", "")
	viper.SetDefault("alipay.server_public_key", "")
	viper.SetDefault("alipay.enable_production", false)
	viper.SetDefault("alipay.encrypt_key", "")

	viper.SetDefault("epay.fwd_secret", "")

	viper.SetDefault("log.console", true)
	viper.SetDefault("log.path", "ocrbench.log")

	// Check if config file exists
	configFile := "config.toml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Config file not found; create with defaults
		fmt.Println("Config file not found, creating default config...")

		// Save default configuration to file
		if err := viper.SafeWriteConfigAs(configFile); err != nil {
			fmt.Printf("ERROR: couldn't write default config file: %v", err)
			panic(err)
		}
	}

	// Read in config file and handle any errors
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("ERROR: couldn't read config file: %v", err)
		panic(err)
	}
}
