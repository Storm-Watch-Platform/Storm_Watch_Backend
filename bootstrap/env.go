// package bootstrap

// import (
// 	"log"

// 	"github.com/spf13/viper"
// )

// type Env struct {
// 	AppEnv                 string `mapstructure:"APP_ENV"`
// 	ServerAddress          string `mapstructure:"SERVER_ADDRESS"`
// 	ContextTimeout         int    `mapstructure:"CONTEXT_TIMEOUT"`
// 	DBHost                 string `mapstructure:"DB_HOST"`
// 	DBPort                 string `mapstructure:"DB_PORT"`
// 	DBUser                 string `mapstructure:"DB_USER"`
// 	DBPass                 string `mapstructure:"DB_PASS"`
// 	DBName                 string `mapstructure:"DB_NAME"`
// 	MongoURI               string `mapstructure:"MONGO_URI"`
// 	AccessTokenExpiryHour  int    `mapstructure:"ACCESS_TOKEN_EXPIRY_HOUR"`
// 	RefreshTokenExpiryHour int    `mapstructure:"REFRESH_TOKEN_EXPIRY_HOUR"`
// 	AccessTokenSecret      string `mapstructure:"ACCESS_TOKEN_SECRET"`
// 	RefreshTokenSecret     string `mapstructure:"REFRESH_TOKEN_SECRET"`
// 	RedisHost              string `mapstructure:"REDIS_HOST"`
// 	RedisPort              string `mapstructure:"REDIS_PORT"`
// 	RedisPass              string `mapstructure:"REDIS_PASS"`
// 	RedisDB                int    `mapstructure:"REDIS_DB"`
// }

// func NewEnv() *Env {
// 	env := Env{}
// 	viper.SetConfigFile(".env")

// 	err := viper.ReadInConfig()
// 	if err != nil {
// 		log.Fatal("Can't find the file .env : ", err)
// 	}

// 	err = viper.Unmarshal(&env)
// 	if err != nil {
// 		log.Fatal("Environment can't be loaded: ", err)
// 	}

// 	if env.AppEnv == "development" {
// 		log.Println("The App is running in development env")
// 	}

// 	return &env
// }

package bootstrap

import (
	"log"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Env struct {
	AppEnv                 string
	ServerAddress          string
	ContextTimeout         int
	DBHost                 string
	DBPort                 string
	DBUser                 string
	DBPass                 string
	DBName                 string
	MongoURI               string
	AccessTokenExpiryHour  int
	RefreshTokenExpiryHour int
	AccessTokenSecret      string
	RefreshTokenSecret     string
	RedisHost              string
	RedisPort              string
	RedisPass              string
	RedisDB                int
	GeminiAPIKey           string
}

func NewEnv() *Env {
	env := &Env{}

	// Cố gắng load .env nếu có (local dev)
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig() // ignore error

	// lấy từ viper hoặc fallback sang os.Getenv
	env.AppEnv = getString("APP_ENV", "production")
	env.ServerAddress = getString("SERVER_ADDRESS", ":8080")
	env.ContextTimeout = getInt("CONTEXT_TIMEOUT", 10)
	env.DBHost = getString("DB_HOST", "localhost")
	env.GeminiAPIKey = getString("GEMINI_API_KEY", "")
	env.DBPort = getString("DB_PORT", "5432")
	env.DBUser = getString("DB_USER", "user")
	env.DBPass = getString("DB_PASS", "pass")
	env.DBName = getString("DB_NAME", "dbname")
	env.MongoURI = getString("MONGO_URI", "")
	env.AccessTokenExpiryHour = getInt("ACCESS_TOKEN_EXPIRY_HOUR", 1)
	env.RefreshTokenExpiryHour = getInt("REFRESH_TOKEN_EXPIRY_HOUR", 24)
	env.AccessTokenSecret = getString("ACCESS_TOKEN_SECRET", "")
	env.RefreshTokenSecret = getString("REFRESH_TOKEN_SECRET", "")
	env.RedisHost = getString("REDIS_HOST", "localhost")
	env.RedisPort = getString("REDIS_PORT", "6379")
	env.RedisPass = getString("REDIS_PASS", "")
	env.RedisDB = getInt("REDIS_DB", 0)

	if env.AppEnv == "development" {
		log.Println("The App is running in development env")
	}

	return env
}

// helper
func getString(key, defaultVal string) string {
	if val := viper.GetString(key); val != "" {
		return val
	}
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getInt(key string, defaultVal int) int {
	if viper.IsSet(key) {
		return viper.GetInt(key)
	}
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
