package httphandler

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gensword/cornmanager"
	"github.com/gin-gonic/gin"
)

func JwtValid() gin.HandlerFunc {
	return func (c *gin.Context) {
		tokenStr := c.Request.Header.Get("Authorization")
		if tokenStr != "" {
			token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (i interface{}, e error) {
				return []byte(cronmanager.Config.GetString("JWT.secretKey")), nil
			})

			if !token.Valid{
				c.JSON(cronmanager.UNAUTH, "not auth user")
				c.Abort()
			} else {
				c.Next()
			}
		} else {
			c.JSON(cronmanager.UNAUTH, "not auth user")
			c.Abort()
		}
	}
}