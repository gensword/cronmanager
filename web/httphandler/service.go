package httphandler

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gensword/cornmanager/client"
	"github.com/gensword/cornmanager/conf"
	"github.com/gensword/cornmanager/cron"
	"github.com/gensword/cornmanager/jobs"
	"github.com/gensword/cornmanager/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"strings"
	"time"
)

// @Summary Register
// @Description add user
// @Tags Register
// @Produce json
// @Param credential body param.User true "username and password"
// @Success 201 {array} httphandler.Response
// @Failure 400 {array} httphandler.Response
// @Router /register [post]
func Register(c *gin.Context) {
	var user model.User
	if err := c.Bind(&user); err != nil {
		c.JSON(conf.BADREQUEST, &Response{Code:conf.BADREQUEST, Message:"bad params"})
		return
	}
	if user.Password == "" || user.UserName == "" {
		c.JSON(conf.BADREQUEST, &Response{Code:conf.BADREQUEST, Message:"username and password are required"})
		return
	}
	if !conf.MsClient.Where(&model.User{UserName:user.UserName}).First(&model.User{}).RecordNotFound() {
		c.JSON(conf.CONFLICT, &Response{Code:conf.CONFLICT, Message:"username already in use"})
		return
	} else {
		hashPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		user.Password = string(hashPassword)
		conf.MsClient.Create(&user)
		c.JSON(conf.CREATED, &Response{Code:conf.CREATED, Message:"register success"})
	}
}

// @Summary JobList
// @Description get jobs list
// @Tags Jobs
// @Produce json
// @Param status query int false "job status(0 stop 1 running)" default(1) Enums(0, 1)
// @Param job_name query string false "job name condition query"
// @Param Authorization header sting false "jwt token for auth"
// @Success 200 {array} httphandler.Response
// @Router /jobs [get]
func GetJobList(c *gin.Context) {
	jobList := client.GetJobs()
	status := c.DefaultQuery("status", "none")
	jobName := c.DefaultQuery("job_name", "none")
	if status == "none" && jobName == "none" {
		c.JSON(conf.SUCCESS, &Response{Code:conf.SUCCESS, Data:jobList})
		return
	} else {
		jobsAfterFilter := make([]*jobs.Job, 0)
		for _, job := range jobList {
			if status != "none" {
				statusInt, _ := strconv.Atoi(status)
				if job.Status == statusInt {
					if jobName == "none" || job.Name == jobName{
						jobsAfterFilter = append(jobsAfterFilter, job)
					}
				}
			} else {
				if jobName == "none" || job.Name == jobName {
					jobsAfterFilter = append(jobsAfterFilter, job)
				}
			}
		}
		c.JSON(conf.SUCCESS, &Response{Code:conf.SUCCESS, Data:jobsAfterFilter})
		return
	}
}

// @Summary Get a Single Job
// @Description get a single job by job id
// @Tags Job
// @Produce json
// @Param Authorization header sting false "jwt token for auth"
// @Param job_id path int true "job id"
// @Success 200 {array} httphandler.Response
// @Failure 404 {array} httphandler.Response
// @Router /jobs/{job_id} [get]
func GetJob(c *gin.Context) {
	job, err := client.GetJob(fmt.Sprintf("jobs%s", strings.TrimLeft(c.Param("job_id"), "/")))
	if err != nil {
		c.JSON(conf.NOTFOUND, err.Error())
		return
	}
	c.JSON(conf.SUCCESS, &Response{Code:conf.SUCCESS, Data:job})
}

// @Summary Add a single job
// @Description add a single job
// @Tags Job
// @Produce json
// @Param Authorization header string false "auth token"
// @Param job body param.AddJob true "single job to add"
// @Success 200 {array} httphandler.Response
// @Router /jobs [post]
func AddJob(c *gin.Context) {
	defer func() {
		err := recover(); if err != nil {
			c.JSON(conf.BADREQUEST, &Response{Code:conf.BADREQUEST, Message:"parse spec failed"})
			return
		}
	}()
	redisClient := conf.RedisClient
	redisClient.Lock.Lock()
	defer redisClient.Lock.Unlock()
	jobId := client.GenJobId()
	var job jobs.Job
	err := c.Bind(&job)
	if err != nil {
		c.JSON(conf.BADREQUEST, &Response{Code:conf.BADREQUEST, Message:"invalid params"})
		return
	}
	job.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	job.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	job.Id = jobId
	cron.MycronList.AddJob(job.Spec, &job, fmt.Sprintf("jobs%d", jobId))
	client.AddJob(job)
	conf.Logger.Info(fmt.Sprintf("add job %s", job.Name))
	c.JSON(conf.SUCCESS, &Response{Code:conf.SUCCESS, Data:job})
}

// @Summary Del a single job
// @Description del a single job
// @Tags Job
// @Produce json
// @Param Authorization header sting false "jwt token for auth"
// @Param job_id path int true "job id"
// @Success 200 {array} httphandler.Response
// @Router /jobs/{job_id} [delete]
func RemoveJob(c *gin.Context) {
	jobIdStr := strings.TrimLeft(c.Param("job_id"), "/")
	jobId, err := strconv.Atoi(jobIdStr)
	if err != nil {
		c.JSON(conf.BADREQUEST, "delete jobs failed")
		return
	}
	job, err := client.GetJob(fmt.Sprintf("jobs%d", jobId))
	if err != nil {
		c.String(conf.NOTFOUND, "job not found")
		return
	}
	cron.MycronList.RmByJobIds([]int{jobId})
	conf.Logger.Info(fmt.Sprintf("remove job %s", job.Name))
	c.JSON(conf.SUCCESS, &Response{Code:conf.SUCCESS, Message:"del success"})
}


// @Summary Modify a single job
// @Description modify a single job
// @Tags Job
// @Produce json
// @Param Authorization header sting false "jwt token for auth"
// @Param job body param.ModifyJob true "modify a single job"
// @Success 200 {array} httphandler.Response
// @Failure 400 {array} httphandler.Response
// @Router /jobs [put]
func ChangeJob(c *gin.Context) {
	var job jobs.Job
	err := c.Bind(&job)
	if err != nil {
		c.JSON(conf.BADREQUEST, &Response{Code:conf.BADREQUEST, Message:"invalid params"})
		return
	}
	originJob, err := client.GetJob(fmt.Sprintf("jobs%d", job.Id))
	if err != nil {
		c.JSON(conf.NOTFOUND, &Response{Code:conf.NOTFOUND, Message:"job not found"})
		return
	}
	job.CreateTime = originJob.CreateTime
	job.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	cron.MycronList.StopJob([]int{originJob.Id})
	//cron.MycronList.RemoveJob(fmt.Sprintf("jobs%d", job.Id))
	defer func(jobId int) {
		if err := recover(); err != nil {
			cron.MycronList.StartJob([]int{jobId})
			c.JSON(conf.BADREQUEST, &Response{Code:conf.BADREQUEST, Message:"parse spec failed"})
		}
	}(originJob.Id)
	if job.Status == 1 {
		cron.MycronList.AddJob(job.Spec, &job, fmt.Sprintf("jobs%d", job.Id))
	}
	client.ChangeJob(originJob.Id, job)
	conf.Logger.Info(fmt.Sprintf("change job %s origin job %+v now job %+V", job.Name, originJob, job))
	c.JSON(conf.SUCCESS, &Response{Code:conf.SUCCESS, Data:job})
}

// @Summary Logs list
// @Description Get logs list
// @Tags Logs
// @Produce json
// @Param Authorization header sting false "jwt token for auth"
// @Param job_id path int false "get logs for a special job"
// @Param status query int false "0 query failed job logs, 1 query success job logs"
// @Param page query int false "page num" default(1)
// @Param limit query int false "page size" default(30)
// @Success 200 {array} httphandler.Response
// @Router /logs/job/{job_id} [get]
func GetLogList(c *gin.Context) {
	var logList [] model.Log
	limitStr := c.DefaultQuery("limit", "30")
	pageStr := c.DefaultQuery("page", "1")
	statusStr := c.DefaultQuery("status", "none")
	jobIdStr := strings.TrimLeft(c.Param("job_id"), "/")
	msClient := conf.MsClient
	limit, _ := strconv.Atoi(limitStr)
	page, _ := strconv.Atoi(pageStr)
	var total, jobId, statusInt int
	if jobIdStr == "" && statusStr == "none"{
		msClient.Table("logs").Count(&total)
	} else {
		if statusStr == "none"{
			jobId, _ = strconv.Atoi(jobIdStr)
			msClient.Model(&model.Log{}).Where("job_id = ?", jobId).Count(&total)
		} else if jobIdStr == ""{
			statusInt, _ = strconv.Atoi(statusStr)
			msClient.Model(&model.Log{}).Where("status = ?",statusInt).Count(&total)
		} else {
			jobId, _ = strconv.Atoi(jobIdStr)
			statusInt, _ = strconv.Atoi(statusStr)
			msClient.Model(&model.Log{}).Where("job_id = ? AND status = ?", jobId, statusInt).Count(&total)
		}
	}
	if limit < 0 {
		limit = 1
	}
	if page < 0 {
		page = 1
	} else if page * limit > total {
		page = total / limit
	}
	start := (page - 1) * limit
	if jobIdStr == "" && statusStr == "none"{
		msClient.Offset(start).Limit(limit).Find(&logList)
	} else if statusStr == "none"{
		msClient.Where("job_id = ?", jobId).Offset(start).Limit(limit).Find(&logList)
	} else if jobIdStr == "" {
		msClient.Where("status = ?", statusInt).Offset(start).Limit(limit).Find(&logList)
	} else {
		msClient.Where("status = ? and job_id = ?", statusInt, jobId).Offset(start).Limit(limit).Find(&logList)
	}
	c.JSON(conf.SUCCESS, &Response{Code:conf.SUCCESS, Data:logList})
}

// @Summary Single Log
// @Description Get a single log
// @Tags Log
// @Produce json
// @Param Authorization header sting false "jwt token for auth"
// @Param log_id path string false "log id"
// @Success 200 {array} httphandler.Response
// @Router /log/{log_id} [get]
func GetLog(c *gin.Context) {
	var log model.Log
	msClient :=  conf.MsClient
	logIdStr := strings.Trim(c.Param("log_id"), " ")
	logId, _:= strconv.Atoi(logIdStr)
	if msClient.Debug().Where("id = ?", logId).First(&log).RecordNotFound() {
		c.String(conf.NOTFOUND, "log not found")
		return
	}
	c.JSON(conf.SUCCESS, &Response{Code:conf.SUCCESS, Data:log})
}

// @Summary Login to get jwt token
// @Description Login to get jwt token
// @Tags Login
// @Param credential body model.User true "username and password"
// @Success 200 {array} httphandler.Response
// @Failure 401 {array} httphandler.Response
// @Router /login [post]
func Login(c *gin.Context) {
	var(
		user model.User
		record model.User
	)
	if err := c.Bind(&user); err != nil {
		c.JSON(conf.UNAUTH, &Response{Code:conf.UNAUTH, Message:"invalid credentials"})
		return
	}
	if conf.MsClient.Where(&model.User{UserName:user.UserName}).First(&record).RecordNotFound() {
		c.JSON(conf.UNAUTH, &Response{Code:conf.UNAUTH, Message:"invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(record.Password), []byte(user.Password)); err != nil {
		c.JSON(conf.UNAUTH, &Response{Code:conf.UNAUTH, Message:"invalid credentials"})
		return
	}
	claim := jwt.MapClaims{
		"id": user.ID,
		"username": user.UserName,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Duration(24) * time.Hour),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenStr, err := token.SignedString([]byte(conf.Config.GetString("JWT.secretKey")))
	if err != nil {
		c.JSON(conf.UNAUTH, &Response{Code:conf.UNAUTH, Message:"invalid credentials"})
		return
	}
	c.JSON(conf.SUCCESS, &Response{Code:conf.SUCCESS, Data:tokenStr})
}
