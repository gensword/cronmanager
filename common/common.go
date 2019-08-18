package common

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pkg/errors"
	"strconv"
	"sync"
)

type redisConf struct {
	Addr     string
	Port     int
	Password string
	Db       int
}

var Rconf redisConf
type RedisClient struct {
	Lock *sync.Mutex
	redis.Client
}

type mysqlConf struct {
	Addr     string
	Port     int
	User     string
	Password string
	DbName   string
	Charset  string
}
type MysqlClient struct {
	gorm.DB
}
var Mconf mysqlConf

func GetRedisClient() (client * RedisClient, err error) {
	newclient := redis.NewClient(&redis.Options{
		Addr:     Rconf.Addr + ":" + strconv.Itoa(Rconf.Port),
		Password: Rconf.Password,
		DB:       Rconf.Db,
	})
	client = &RedisClient{new(sync.Mutex), *newclient}
	if _, err := client.Ping().Result(); err != nil {
		redisErr := errors.Wrap(err, "redis client pong failed")
		return client, redisErr
	}
	return client, nil
}

func GetMysqlClient() (client *MysqlClient, err error) {
	host := fmt.Sprintf("%s:%d", Mconf.Addr, Mconf.Port)
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=True&loc=Local", Mconf.User, Mconf.Password, host, Mconf.DbName, Mconf.Charset))
	if err != nil {
		err = errors.Wrap(err, "connect mysql failed")
	}
	client = &MysqlClient{*db}
	return client, err
}