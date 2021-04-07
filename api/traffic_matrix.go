package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"strconv"
)

var trafficHandler *redis.Client
var topo [][]bool

func getTrafficMatrix(c *gin.Context)  {
	ctx:=context.Background()
	keys,err:=trafficHandler.Keys(ctx,"*").Result()
	if err!=nil{
		c.JSON(http.StatusInternalServerError,internalErrorJSON)
		return
	}
	res:=make(map[string]float64)
	for i:=0;i<len(topo);i++{
		for j:=0;j<len(topo);j++{
			if i==j{
				continue
			}
			k:=fmt.Sprintf("%d-%d",i,j)
			res[k]=0
		}
	}
	for _,key:=range keys{
		vals,err:=trafficHandler.ZRevRangeByScore(ctx,key,&redis.ZRangeBy{
			Min: "-inf",
			Max: "+inf",
			Count: 1,
		}).Result()
		if err!=nil{
			c.JSON(http.StatusInternalServerError,internalErrorJSON)
			return
		}
		if len(vals)!=1{
			log.Println(len(vals))
			log.Printf("warn: find nothing of key %s\n",key)
			continue
		}
		traffic,err:=strconv.ParseFloat(vals[0],64)
		res[key]=traffic
	}
	// c.JSON(http.StatusOK,gin.H{
	// 	// "res":res,
	// })
	c.JSON(http.StatusOK, res)
	return

}

