package config

import (
	"os"
	"strconv"
)

type Config struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDb       string

	PaymentApiUrl         string
	PaymentApiFallbackUrl string
}

var CONFIG Config

func LoadConfig() {
	CONFIG = Config{
		RedisHost:     getEnv("REDIS_HOST", "rinha-de-backend-2025-redis-1"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDb:       getEnv("REDIS_DB", "0"),

		PaymentApiUrl:         getEnv("PAYMENT_API_URL", "http://payment-processor-default:8080"),
		PaymentApiFallbackUrl: getEnv("PAYMENT_API_FALLBACK_URL", "http://payment-processor-fallback:8080"),
	}

}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if valStr == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultVal
	}
	return b
}
