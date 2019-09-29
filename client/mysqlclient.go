package client

import (
	"github.com/gensword/cornmanager"
	"github.com/gensword/cornmanager/jobs"
	"github.com/gensword/cornmanager/model"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func CloseClient() {
	cronmanager.MsClient.Close()
}

func  Log() {
	for log := range jobs.LogChan {
		newLog := &model.Log{JobId:log.JobId, JobName:log.JobNmae, OutPut:log.Message, StartTime:log.StartTime, EndTime:log.EndTime, Status:log.IsSuccess}
		cronmanager.MsClient.Create(newLog)
	}
}
