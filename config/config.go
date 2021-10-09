package config

type AppConfig struct {
	AppName        string `env:"APP_NAME" env-default:"MyProduct"`
	AppEnv         string `env:"APP_ENV" env-default:"DEV"`
	Port           string `env:"Port" env-default:"8081"`
	Host           string `env:"HOST" env-default:"localhost"`
	LogLevel       string `env:"LOG_LEVEL" env-default:"ERROR"`
	DBHost         string `env:"DB_HOST" env-default:"localhost"`
	DBPort         string `env:"DB_PORT" env-default:"27017"`
	DBName         string `env:"DB_NAME" env-default:"sample_product"`
	CollectionName string `env:"COLLECTION_NAME" env-default:"products"`
	UserCollection string `env:"USER_COL_NAME" env-default:"users"`
	JwtTokenSecret string `env:"JWT_TOKEN_SECRET" env-default:"token"`
}

var cfg AppConfig
