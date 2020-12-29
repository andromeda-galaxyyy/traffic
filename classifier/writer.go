package main

import (
	"chandler.com/gogen/utils"
	"github.com/go-redis/redis/v8"
)

type redisWriter struct {
	ip string
	port int
	handle *redis.Client
}

func newDefaultRedisWriter(ip string,port int) *redisWriter {
	return &redisWriter{
		ip:     ip,
		port:   port,
		handle: nil,
	}
}

func (writer *redisWriter)Init()(err error)  {
	writer.handle,err=utils.NewRedisClient(writer.ip,writer.port,6)
	if err!=nil{

	}
	return nil
}

func (writer *redisWriter)Write(r *result) error{
	return nil
	//ctx:=context.Background()
	//if err:
}


