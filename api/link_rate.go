package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetLinkRate(context *gin.Context)  {
	context.JSON(http.StatusOK,gin.H{
		"msg":"ok",
	})
	return
}