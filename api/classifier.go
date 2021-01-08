package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
)

var (
	classifierRedisHandle *redis.Client
)

func getClassifierTestCase(c *gin.Context){
	from:=c.GetInt64("from")
	to:=c.GetInt64("to")
	log.Println(from)
	log.Println(to)
	c.JSON(http.StatusOK,gin.H{
		"msg":"hello world",
	})
	return
}
