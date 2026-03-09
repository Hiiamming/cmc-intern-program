package config

import "github.com/spf13/viper"

type PostgresConfig struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
}

func LoadPostgresConfig(file_name string) (config *PostgresConfig, err error) {
	viper.SetConfigFile(file_name)

	// Set default values
	viper.SetDefault("DB_NAME", "mini_asm")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	config = &PostgresConfig{}
	err = viper.Unmarshal(config)
	if err != nil {
		return
	}

	return config, err
}
