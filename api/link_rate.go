package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"net/http"
	"strconv"
)

var linkRateRedisHandle *redis.Client

func init()  {
	//var err error
	//linkRateRedisHandle,err=utils.NewRedisClient("10.211.55.1",6379,6)
	//if err!=nil{
	//	log.Fatalln(err)
	//}

}

func LinkRateDemo(c *gin.Context)  {
	c.JSON(http.StatusOK,gin.H{
		"msg":c.GetString("hello"),
	})
	return
}
// /linkrate?src=&dst=&from=&to=&
func GetLinkRate(c *gin.Context)  {
	src,ok:=c.GetQuery("src")
	if !ok{
		c.JSON(http.StatusBadRequest,invalidRequestJSON)
		return
	}

	dst,ok:=c.GetQuery("dst")
	if !ok{
		c.JSON(http.StatusBadRequest,invalidRequestJSON)
		return
	}

	from:=c.GetInt64("from")
	to:=c.GetInt64("to")
	key:=fmt.Sprintf("%s-%s",src,dst)
	vals,err:=linkRateRedisHandle.ZRangeByScore(context.Background(),key,&redis.ZRangeBy{
		Min: fmt.Sprintf("%d", from),
		Max: fmt.Sprintf("%d", to),
	}).Result()
	if err!=nil{
		c.JSON(http.StatusInternalServerError,internalErrorJSON)
		return
	}
	if len(vals)==0{
		c.JSON(http.StatusNotFound,notFoundJSON)
		return
	}
	res:=make([]interface{},0)
	for _,val:=range vals{
		r,err:=strconv.ParseFloat(val,64)
		if err!=nil{
			c.JSON(http.StatusInternalServerError,internalErrorJSON)
			return
		}
		res=append(res,r)
	}
	c.JSON(http.StatusOK,gin.H{
		"count":len(res),
		"data":res,
	})
	return
}

func getMaxRate(c *gin.Context)  {
	// get all keys
	ctx:=context.Background()
	var res int=0
	//var cursor uint64
	for {
		var keys []string
		var err error
		//keys,cursor,err=linkRateRedisHandle.Scan(ctx,cursor,"",10).Result()

		keys,err=linkRateRedisHandle.Keys(ctx,"*").Result()
		if err!=nil{
			c.JSON(http.StatusInternalServerError,internalErrorJSON)
			return
		}


		for _,key:=range keys{
			//log.Println(key)
			vals,err:=linkRateRedisHandle.ZRevRangeByScore(ctx,key,&redis.ZRangeBy{
				Min: "-inf",
				Max: "+inf",
				Count: 1,
			}).Result()
			if err!=nil{
				c.JSON(http.StatusInternalServerError,internalErrorJSON)
				return
			}
			maxRate,err:=strconv.Atoi(vals[0])
			if err!=nil{
				c.JSON(http.StatusInternalServerError,internalErrorJSON)
				return
			}
			if maxRate>res{
				res=maxRate
			}
		}
		c.JSON(http.StatusOK,gin.H{
			"res":res,
		})
		return
	}
}