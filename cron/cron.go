package cron

import (
	"fmt"
	"github.com/gensword/cornmanager/client"
	"github.com/gensword/cornmanager/conf"
	"github.com/gensword/cornmanager/jobs"
	"github.com/gensword/cornmanager/model"
	"github.com/jakecoffman/cron"
)

type MyCron struct {
	cron.Cron
}

var MycronList MyCron = MyCron{
	*cron.New(),
}

func (myCron MyCron) RmByJobIds(jobIds []int) error {
	mysqlClient := conf.MsClient
	for _, jobId := range jobIds {
		myCron.RemoveJob(fmt.Sprintf("jobs%d", jobId))
		client.RemoveJob(jobId)
		mysqlClient.Where("job_id = ?", jobId).Delete(&model.Log{})
	}
	return nil
}

func InitCrons() error {
	for _, job := range client.GetJobs() {
		if job.Status == 1 {
			MycronList.AddJob(job.Spec, job, fmt.Sprintf("jobs%d", job.Id))
		}
	}
	return nil
}

func (mycron MyCron) StopJob(jobIds []int) error {
	redisClient := conf.RedisClient
	for _, jobId := range jobIds {
		mycron.RemoveJob(fmt.Sprintf("jobs%d", jobId))
		_, err := redisClient.HSet(fmt.Sprintf("jobs%d", jobId), "Status", 0).Result()
		if err != nil {
			return nil
		}
	}
	return nil
}

func (mycron MyCron) StartJob(jobIds []int) error {
	redisClient := conf.RedisClient
	for _, jobId := range jobIds {
		jobFields, err := redisClient.HGetAll(fmt.Sprintf("jobs%d", jobId)).Result()
		if err != nil {
			return err
		}
		job := &jobs.Job{
			//Id:         id,
			Name:       jobFields["Name"],
			Cmd:        jobFields["Cmd"],
			Spec:       jobFields["Spec"],
		}
		mycron.AddJob(job.Spec, job, fmt.Sprintf("jobs%d", jobId))
		_, err = redisClient.HSet(fmt.Sprintf("jobs%d", jobId), "Status", 1).Result()
		if err != nil {
			return err
		}
	}
	return nil
}
