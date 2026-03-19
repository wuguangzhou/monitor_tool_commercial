package config

type ThirdApiConfig struct {
	EmailHost      string
	EmailPort      int
	EmailUsername  string
	EmailPassword  string
	WechatPayAppId string
}

var thirdApiConfig ThirdApiConfig

// 初始化第三方API配置（后续对接时完善）
func InitThirdApiConfig() {
	thirdApiConfig = ThirdApiConfig{
		// 暂时留空，Day5对接邮件、Day6对接支付时补充
	}
}
