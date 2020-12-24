package utils

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

func NewRedisClient(ip string,port int,db int) (*redis.Client,error){
	client:=redis.NewClient(&redis.Options{
		DialTimeout: 2*time.Second,
		MaxRetries: 1,
		Addr: fmt.Sprintf("%s:%d",ip,port),
		DB: db,
	})
	_,err:=client.Ping(context.Background()).Result()
	if err!=nil{
		return nil,err
	}
	return client,nil
}
