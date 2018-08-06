package controllers

import (
	"github.com/gin-gonic/gin"
	"qiniupkg.com/api.v7/auth/qbox"
)

func VerifyAuth(c *gin.Context) {

	mac := qbox.NewMac(accessKey, secretKey)
	res, err := mac.VerifyCallback(c.Request)
	if err == nil && res == true {
		c.JSON(200, nil)
	} else {
		c.JSON(401, gin.H{"error": err})
	}
}
