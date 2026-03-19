package config

import (
	"log"

	"github.com/joho/godotenv"
)

// 初始化配置
func InitConfig() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Println("未找到.env文件，使用系统变量")
	}

	// 初始化数据库配置
	InitDatabaseConfig()

	// 初始化redis配置
	InitRedisConfig()

	// 初始化第三方API配置
	InitThirdApiConfig()
}
