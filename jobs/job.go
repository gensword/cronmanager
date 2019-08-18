package jobs

import (
	"os/exec"
	"time"
	"fmt"
)

type Job struct {
	Id         int    `json:"id" form:"id"`
	Name       string `json:"name" form:"name"`
	Cmd        string `json:"command" form:"command"`
	Spec       string `json:"spec" form:"spec"`
	Status     int    `json:"status" form:"status"`      // 0 暂停，1 正常
	CreateTime string `json:"create_time" form:"create_time"`
	UpdateTime string `json:"update_time" form:"update_time"`
}

type LogRecord struct {
	JobId     int
	JobNmae   string
	IsSuccess int
	Message   string
	StartTime time.Time
	EndTime   time.Time
}

var LogChan chan *LogRecord = make(chan *LogRecord, 100)

func NewJob() *Job {
	return &Job{}
}

func (job *Job) Run() {
	startTime := time.Now()
	command := job.Cmd
	cmd := exec.Command("/bin/sh", "-c", command)
	output, err := cmd.Output()
	fmt.Println(string(output))
	endTime := time.Now()
	if err != nil {
		LogChan <- &LogRecord{JobId: job.Id, IsSuccess: 0, JobNmae:job.Name, Message: err.Error(), StartTime: startTime, EndTime: endTime, }
	} else {
		LogChan <- &LogRecord{JobId: job.Id, IsSuccess: 1, JobNmae:job.Name, Message: string(output), StartTime: startTime, EndTime: endTime}
	}
}
