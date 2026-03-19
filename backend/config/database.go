package config

import (
	"os"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Dbname   string
	Charset  string
}

var DBConfig DatabaseConfig

func InitDatabaseConfig() {
	//port,_:=strconv.Atoi(getEnv("DB_PORT","3306"))
	DBConfig = DatabaseConfig{
		Host:     getEnv("DB_HOST", "127.0.0.1"),
		Port:     getEnv("DB_PORT", "3306"),
		Username: getEnv("DB_USERNAME", "root"),
		Password: getEnv("DB_PASSWORD", "123456"),
		Dbname:   getEnv("DB_NAME", ""),
		Charset:  getEnv("DB_CHARSET", "utf8mb4"),
	}
}

// 辅助函数
func getEnv(key, defaultVal string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return value
}
