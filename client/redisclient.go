package client

import (
	"fmt"
	"github.com/gensword/cornmanager/conf"
	"github.com/gensword/cornmanager/jobs"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
)

func ChangeJob(jobId int, job jobs.Job) error{
	if exist, _ := conf.RedisClient.Exists(fmt.Sprintf("jobs%d", jobId)).Result(); exist == 0 {
		return errors.New(fmt.Sprintf("jobs%d not exist", jobId))
	}
	t := reflect.TypeOf(job)
	v := reflect.ValueOf(job)
	fields := make(map[string]interface{}, t.NumField())
	for k := 0; k < t.NumField(); k ++ {
		fields[t.Field(k).Name] = v.Field(k).Interface()
	}
	conf.RedisClient.HMSet(fmt.Sprintf("jobs%d", job.Id), fields).Result()
	return nil
}

func AddJob(job jobs.Job) error{
	t := reflect.TypeOf(job)
	v := reflect.ValueOf(job)
	fields := make(map[string]interface{}, t.NumField())
	for k := 0; k < t.NumField(); k ++ {
		fields[t.Field(k).Name] = v.Field(k).Interface()
	}
	_, err := conf.RedisClient.HMSet(fmt.Sprintf("jobs%d", job.Id), fields).Result()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("add job %+v failed", job))
	}
	conf.RedisClient.Incr("lastId")
	return nil
}

func RemoveJob(jobId int) error{
	jobsKey := fmt.Sprintf("jobs%d", jobId)
	_, err := conf.RedisClient.Del(jobsKey).Result()
	return err
}

func GetJobs()(jobs []*jobs.Job) {
	var cursor uint64
	firstLoop := true
	for keys, nextCursor, _ := conf.RedisClient.Scan(cursor, "jobs*", 1000).Result(); nextCursor != 0 || firstLoop; {
		firstLoop = false
		for _, key := range keys {
			job, _ := GetJob(key)
			jobs = append(jobs, job)
		}
		cursor = nextCursor
	}
	return jobs
}

func GetJob(key string)(job *jobs.Job, err error) {
	if exist, _ := conf.RedisClient.Exists(key).Result(); exist == 0 {
		return nil, errors.New("job not found")
	}
	jobFields, err := conf.RedisClient.HGetAll(key).Result()
	id, _ := strconv.Atoi(jobFields["Id"])
	status, _ := strconv.Atoi(jobFields["Status"])
	job = &jobs.Job{
		Id: id,
		Name: jobFields["Name"],
		Cmd: jobFields["Cmd"],
		Spec: jobFields["Spec"],
		Status: status,
		CreateTime: jobFields["CreateTime"],
		UpdateTime: jobFields["UpdateTime"],
	}
	return job, nil
}

func GenJobId()(jobId int) {
	prevJobIdString, _ := conf.RedisClient.Get("lastId").Result()
	prevJobId, _ := strconv.Atoi(prevJobIdString)
	return prevJobId + 1
}