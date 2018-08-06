package controllers

import "github.com/gin-gonic/gin"

func Login(c *gin.Context) {

	c.HTML(200, "index.html", nil)
}
