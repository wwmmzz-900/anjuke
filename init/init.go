package init

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB
var RDB *redis.Client

func init() {
	MysqlInit()
	RedisInit()
}

/*
	func ViperInit() {
		v := viper.New()
		v.SetConfigFile("configs/config.yaml")

		v.ReadInConfig()

}
*/
func MysqlInit() {
	var err error
	dsn := "root:e10adc3949ba59abbe56e057f20f883e@tcp(14.103.149.201:3306)/anjuke?parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("mysql 连接失败")
	}
}

func RedisInit() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "14.103.149.201:6379",
		Password: "e10adc3949ba59abbe56e057f20f883e", // no password set
		DB:       0,                                  // use default DB
	})

}
