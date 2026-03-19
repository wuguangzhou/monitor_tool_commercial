package config

import (
	"strconv"
)

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	Db       int
}

var RDConfig RedisConfig

func InitRedisConfig() {
	port, _ := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	db, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	RDConfig = RedisConfig{
		Host:     getEnv("REDIS_HOST", "127.0.0.1"),
		Port:     port,
		Password: getEnv("REDIS_PASSWORD", "123456"),
		Db:       db,
	}
}
