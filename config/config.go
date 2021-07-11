package config

import "os"

var (
	RedisHost string
)

func SetConfig() {
	RedisHost = os.Getenv("REDIS_URL")
}
