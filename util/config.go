package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration vatiables for the application
// that we read from file or env. variables
type Config struct {
	Environment          string        `mapstructure:"ENVIROMENT"`
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	MIGRATION_URL        string        `mapstructure:"MIGRATION_URL"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

// read configuration from the file using the path if file exist for development
// or overrifde them with env. variables for production using docker
func LoadConfig(path string) (config Config, err error) {
	// set the location of the config file (app.env)
	viper.AddConfigPath(path)
	// sets name for the config file without extension
	viper.SetConfigName("app")
	// sets the type of the configuration returned by the remote source , extension
	viper.SetConfigType("env")

	// makes Viper check if environment variables match any of the existing keys
	// if exist , override the app.env variables
	viper.AutomaticEnv()

	// read a configuration file, setting existing keys to nil if the key does not exist in the file
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	return
}
