package controllers

import "github.com/gin-gonic/gin"

func Index(c *gin.Context) {

	c.HTML(200, "overview.html", nil)
}
