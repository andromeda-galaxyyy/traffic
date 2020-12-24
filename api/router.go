package main

import (
	"chandler.com/gogen/models"
	"github.com/gin-gonic/gin"
	// "github.com/go-redis/redis/v8"
)

var (
	router *gin.Engine
	invalidRequestJSON=gin.H{
		"msg":"invalid request",
	}
	internalErrorJSON=gin.H{
		"msg":"internal error",
	}
	notFoundJSON=gin.H{
		"msg":"not found",
	}

	counterReader *models.FCounterReader
)



func GinHelloWorld(context *gin.Context)  {
	context.JSON(200,gin.H{
		"message":"helloworld",
	})
}


