package main

import (
	"chandler.com/gogen/models"
	"context"
	"fmt"
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
	vals,err:=classifierRedisHandle.ZRangeByScore(context.Background(),"classifier_test",&redis.ZRangeBy{
		Min: fmt.Sprintf("%d",from),
		Max:fmt.Sprintf("%d",to),
	}).Result()
	if err!=nil{
		c.JSON(http.StatusInternalServerError,internalErrorJSON)
		return
	}
	tmpRes :=make([]*models.TestStats,0)
	for _,val:=range vals{
		testCase:=&models.TestStats{}
		err=testCase.UnBox([]byte(val))
		if err!=nil{
			c.JSON(http.StatusInternalServerError,internalErrorJSON)
			return
		}
		tmpRes =append(tmpRes,testCase)
	}
	if len(tmpRes)==0{
		c.JSON(http.StatusNotFound,notFoundJSON)
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"num":  len(tmpRes),
		"data": tmpRes,
	})
	return
}
