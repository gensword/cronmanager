package main

import (
	"flag"
	"fmt"
	"github.com/gensword/cornmanager"
	"github.com/gensword/cornmanager/client"
	"github.com/gensword/cornmanager/cron"
	"github.com/gensword/cornmanager/model"
	"github.com/gensword/cornmanager/web/httphandler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// @title Cronmanager Api
// @version 1.0
// @description a cron manage instead of linux crontab
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api/v1
func main() {
	shouldCreateTable := flag.Bool("create_table", false, "should create table")
	flag.Parse()
	if *shouldCreateTable {
		cronmanager.MsClient.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&model.User{})
		cronmanager.MsClient.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&model.Log{})
	}

	cron.InitCrons()
	cron.MycronList.Start()

	router := httphandler.GetRouter()
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cronmanager.Config.GetInt("httpServer.port")),
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	go func() {
		client.Log()
	}()
	cronmanager.Logger.Info("server start")
	done := make(chan bool)
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<- sigs
		done <- true
	}()

	<- done

	srv.Close()
	cron.MycronList.Stop()
	client.CloseClient()
	cronmanager.RedisClient.Close()
	cronmanager.Logger.Info("server shutdown normally")
}