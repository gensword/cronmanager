package cronmanager

import (
	"fmt"
	"github.com/gensword/cornmanager/common"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"path"
)

const (
	SUCCESS = 200
	BADREQUEST = 400
	CREATED = 201
	UNAUTH = 401
	CONFLICT = 409
	NOTFOUND = 404
)

var Config *viper.Viper
var Logger *zap.Logger
var MsClient *common.MysqlClient
var RedisClient *common.RedisClient

func init() {
	Config = viper.New()
	Config.SetConfigName("config")
	Config.AddConfigPath("/Users/mac/cronmanager/conf/")
	Config.SetConfigType("yaml")
	err := Config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	err = Config.UnmarshalKey("redis", &common.Rconf)
	if err != nil {
		panic("redis conf wrong")
	}
	err = Config.UnmarshalKey("mysql", &common.Mconf)
	Logger, _ = NewLogger()
	MsClient, err = common.GetMysqlClient()
	if err != nil {
		panic(err)
	}
	RedisClient, err = common.GetRedisClient()
	if err != nil {
		panic(err)
	}
}

func NewLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		path.Join(Config.GetString("log.path"), Config.GetString("log.fileName")), "stdout",
	}
	return cfg.Build()
}
