package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	telemetryHandle *redis.Client
)

func TelemetryFunc(isDelay bool) func(c *gin.Context)  {
	var keyTemplate string
	if isDelay{
		keyTemplate="%s-%s.delay"
	}else{
		keyTemplate="%s-%s.loss"
	}
	return func(c *gin.Context) {
		u,ok:=c.GetQuery("u")
		if !ok{
			c.JSON(http.StatusBadRequest, invalidRequestJSON)
			return
		}
		v,ok:=c.GetQuery("v")
		if !ok{
			c.JSON(http.StatusBadRequest,invalidRequestJSON)
			return
		}
		var count int64=5
		countStr,ok:=c.GetQuery("count")
		if ok{
			co,err:=strconv.ParseInt(countStr,10,64)
			if err!=nil{
				c.JSON(http.StatusBadRequest,invalidRequestJSON)
				return
			}
			count=co
		}

		key:=fmt.Sprintf(keyTemplate,u,v)
		ctx:=context.Background()
		delays,err:=telemetryHandle.ZRangeByScore(ctx,key,&redis.ZRangeBy{
			Min: "-inf",
			Max:"+inf",
			Count: count,
		}).Result()

		if err!=nil{
			c.JSON(http.StatusInternalServerError,internalErrorJSON)
			return
		}
		res:=make([]float64,0)
		for _,delayStr:=range delays{
			d,err:=strconv.ParseFloat(delayStr,64)
			if err!=nil{
				continue
			}
			res=append(res,d)
		}
		c.JSON(http.StatusOK, gin.H{
			"count":len(res),
			"res":res,
		})
		return
	}
}

// get stats of all link
func TelemetryFunc2(isDelay bool) func(c *gin.Context) {
	var keyTemplate string
	if isDelay{
		keyTemplate="*.delay"
	}else{
		keyTemplate="*.loss"
	}
	return func(c *gin.Context) {
		res:=make(map[string]float64)
		for i:=0;i<100;i++{
			for j:=0;j<100;j++{
				if i>=j{
					continue
				}
				if !topo[i][j]{
					continue
				}
				k:=fmt.Sprintf("%d-%d",i,j)
				res[k]=0
			}
		}
		ctx:=context.Background()
		keys,err:=telemetryHandle.Keys(ctx,keyTemplate).Result()
		if err!=nil{
			c.JSON(http.StatusInternalServerError,internalErrorJSON)
			return
		}
		for _,key:=range keys{
			// 1-2.delay
			//1-2.loss
			//1-2
			kk:=strings.Split(key,".")[0]
			//fix if we meet 2-1
			nodesStr:=strings.Split(kk,"-")
			if len(nodesStr)!=2{
				c.JSON(http.StatusInternalServerError,internalErrorJSON)
				return
			}
			aa,bb:=nodesStr[0],nodesStr[1]
			a,err:=strconv.Atoi(aa)
			if err!=nil{
				c.JSON(http.StatusInternalServerError,internalErrorJSON)
				return
			}
			b,err:=strconv.Atoi(bb)
			if err!=nil{
				c.JSON(http.StatusInternalServerError,internalErrorJSON)
				return
			}
			if a>b{
				kk=fmt.Sprintf("%d-%d",b,a)
			}
			vals,err:=telemetryHandle.ZRevRangeByScore(ctx,key,&redis.ZRangeBy{
				Min: "-inf",
				Max:"+inf",
				Count: 1,
			}).Result()
			if err!=nil{
				c.JSON(http.StatusInternalServerError,internalErrorJSON)
				return
			}
			if len(vals)!=1{
				log.Println(len(vals))
				log.Printf("warn: cannot find value of key:%s\n",key)
			}
			rate,err:=strconv.ParseFloat(vals[0],64)
			if err!=nil{
				c.JSON(http.StatusInternalServerError,internalErrorJSON)
				return
			}
			res[kk]=rate
		}
		c.JSON(http.StatusOK,res)
		return
	}
}



//
//func GetLinkDelay(c *gin.Context){
//	u,ok:=c.GetQuery("u")
//	if !ok{
//		c.JSON(http.StatusBadRequest, invalidRequestJSON)
//		return
//	}
//	v,ok:=c.GetQuery("v")
//	if !ok{
//		c.JSON(http.StatusBadRequest,invalidRequestJSON)
//		return
//	}
//	var count int64=5
//	countStr,ok:=c.GetQuery("count")
//	if ok{
//		co,err:=strconv.ParseInt(countStr,10,64)
//		if err!=nil{
//			c.JSON(http.StatusBadRequest,invalidRequestJSON)
//			return
//		}
//		count=co
//	}
//
//	key:=fmt.Sprintf("%s-%s.delay",u,v)
//	ctx:=context.Background()
//	delays,err:=telemetryHandle.ZRangeByScore(ctx,key,&redis.ZRangeBy{
//		Min: "-inf",
//		Max:"+inf",
//		Count: count,
//	}).Result()
//
//	if err!=nil{
//		c.JSON(http.StatusInternalServerError,internalErrorJSON)
//		return
//	}
//	res:=make([]float64,0)
//	for _,delayStr:=range delays{
//		d,err:=strconv.ParseFloat(delayStr,64)
//		if err!=nil{
//			continue
//		}
//		res=append(res,d)
//	}
//	c.JSON(http.StatusOK, gin.H{
//		"count":len(res),
//		"res":res,
//	})
//	return
//}
//
//func GetLinkLoss(context *gin.Context)  {
//	context.JSON(http.StatusOK,gin.H{
//		"msg":"link loss",
//	})
//	return
//}


