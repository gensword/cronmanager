package httphandler

import (
	"github.com/gensword/cornmanager/conf"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"path"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	_"github.com/gensword/cornmanager/docs"
	"time"
)

func GetRouter () *gin.Engine{
	gin.DisableConsoleColor()
	ginLog, err := os.Create(path.Join(conf.Config.GetString("log.path"), conf.Config.GetString("log.ginLog")))
	if err != nil {
		panic(err)
	}
	gin.DefaultWriter = io.MultiWriter(ginLog)
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"PUT", "PATCH", "DELETE", "POST", "GET"},
		AllowHeaders:     []string{"Origin", "authorization", "content-type"},
		ExposeHeaders:    []string{"Content-Length", "authorization"},
		AllowCredentials: true,
		AllowAllOrigins:true,
		MaxAge: 12 * time.Hour,
	}))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	v1 := router.Group("/api/v1")
	{
		// jobs
		v1.GET("/jobs", JwtValid(), GetJobList)
		v1.GET("/jobs/:job_id",JwtValid(), GetJob)
		v1.POST("/jobs", JwtValid(), AddJob)
		v1.PUT("/jobs", JwtValid(), ChangeJob)
		v1.DELETE("/jobs/:job_id", JwtValid(), RemoveJob)

		// log
		v1.GET("/logs/job/*job_id",JwtValid(), GetLogList)
		v1.GET("/log/:log_id", JwtValid(), GetLog)

		// login
		v1.POST("/login", Login)

		// register
		v1.POST("/register", Register)
	}

	return router
}