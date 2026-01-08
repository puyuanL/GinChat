package main

import (
	"GinChat/models"
	"GinChat/router" //  router "GinChat/router"
	"GinChat/utils"
	"time"

	"github.com/spf13/viper"
)

func main() {
	utils.InitConfig()
	utils.InitMySQL()
	//utils.InitRedis()
	utils.InitRedisCluster()
	InitTimer()
	r := router.Router()                  // router.Router()
	r.Run(viper.GetString("port.server")) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

// InitTimer 初始化定时器
func InitTimer() {
	wrapCleanConn := func(param interface{}) bool {
		return models.CleanConnection()
	}
	utils.Timer(
		time.Duration(viper.GetInt("timeout.DelayHeartbeat"))*time.Second,
		time.Duration(viper.GetInt("timeout.HeartbeatHz"))*time.Second,
		wrapCleanConn,
		"",
	)
}
