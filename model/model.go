package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Log struct {
	gorm.Model
	JobId int `gorm:"index:job_id"`
	JobName string
	OutPut string
	Status int // 0 失败 1 成功
	StartTime time.Time
	EndTime time.Time
}

type User struct {
	gorm.Model
	UserName string `json:"user_name,omitempty"gorm:"user_name"`
	Password string `json:"password"gorm:"password"`
}