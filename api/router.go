package main

import (
	"chandler.com/gogen/models"
	"github.com/gin-gonic/gin"
	// "github.com/go-redis/redis/v8"
)

var (
	router *gin.Engine
	invalidRequestJSON=gin.H{
		"message":"invalid request",
	}
	internalErrorJSON=gin.H{
		"message":"internal error",
	}

	counterReader *models.FCounterReader
)



func GinHelloWorld(context *gin.Context)  {
	context.JSON(200,gin.H{
		"message":"helloworld",
	})
}


