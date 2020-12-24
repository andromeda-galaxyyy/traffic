package main

import (
	"chandler.com/gogen/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func TimestampComplete() gin.HandlerFunc  {
	return func(c *gin.Context) {
		var from int64
		var to int64
		var err error
		toStr,ok:=c.GetQuery("to")
		if !ok{
			to=utils.NowInMilli()
		}else{
			to,err=strconv.ParseInt(toStr,10,64)
			if err!=nil{
				c.JSON(http.StatusBadRequest,invalidRequestJSON)
				return
			}
		}


		fromStr,ok:= c.GetQuery("from")
		if !ok{
			from=to-5
		}else{
			from,err=strconv.ParseInt(fromStr,10,64)
			if err!=nil{
				c.JSON(http.StatusBadRequest,invalidRequestJSON)
				return
			}
		}
		c.Set("from",from)
		c.Set("to",to)
		c.Next()
		return
	}
}

