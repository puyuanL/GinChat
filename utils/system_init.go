package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB *gorm.DB
	// RedisCluster RedisClient  *redis.Client
	RedisCluster *redis.ClusterClient
)

func InitConfig() {
	viper.SetConfigName("app")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("config app initialized.")
}

func InitMySQL() {
	//自定义日志模板 打印SQL语句
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, //慢SQL阈值
			LogLevel:      logger.Info, //级别
			Colorful:      true,        //彩色
		},
	)

	DB, _ = gorm.Open(mysql.Open(viper.GetString("mysql.dns")),
		&gorm.Config{Logger: newLogger})
	fmt.Println(" MySQL initialized.")
	//user := models.UserBasic{}
	//DB.Find(&user)
	//fmt.Println(user)
}

//// InitRedis 初始化单节点Redis
//func InitRedis() {
//	RedisClient = redis.NewClient(&redis.Options{
//		Addr:         viper.GetString("redis.addr"),
//		Password:     viper.GetString("redis.password"),
//		DB:           viper.GetInt("redis.DB"),
//		PoolSize:     viper.GetInt("redis.poolSize"),
//		MinIdleConns: viper.GetInt("redis.minIdleConn"),
//	})
//}

// InitRedisCluster 初始化单节点Redis
func InitRedisCluster() {
	RedisCluster = redis.NewClusterClient(&redis.ClusterOptions{
		// 虚拟机 Redis 集群节点列表（必填）
		Addrs:    viper.GetStringSlice("redis.cluster.addrs"),
		Password: "", // 集群密码（若无则留空）
		// 可选配置（根据需要调整）
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读超时
		WriteTimeout: 3 * time.Second, // 写超时
		MaxRedirects: 3,               // 集群重定向最大次数
	})
	fmt.Println(" Redis initialized.", viper.GetStringSlice("redis.cluster.addrs"))
}

const (
	PublishKey = "websocket"
)

// Publish 发布消息到Redis
func Publish(ctx context.Context, channel string, msg string) error {
	var err error
	fmt.Println("Publish...", msg)
	err = RedisCluster.Publish(ctx, channel, msg).Err()
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// Subscribe 订阅Redis消息
func Subscribe(ctx context.Context, channel string) (string, error) {
	sub := RedisCluster.Subscribe(ctx, channel)
	fmt.Println("Subscribe...", ctx)
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println("Subscribe...", msg.Payload)
	return msg.Payload, err
}
