package main

import (
	"chandler.com/gogen/common"
	"chandler.com/gogen/models"
	"chandler.com/gogen/utils"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"strings"
)

type LinkRateHandler struct {
	ip string
	port int
	redis *redis.Client
	done chan common.Signal
	cache chan string
	removeFile bool
}

func (handler *LinkRateHandler) StoreSingleRate(rate *models.Rate,weight int64) error {
	key:=fmt.Sprintf("%s-%s",rate.Src,rate.Dst)
	if err:=handler.redis.ZAdd(
		context.Background(),
		key,
		&redis.Z{
			Score: float64(weight),
			Member: rate.Volume,
		},
		).Err();err!=nil{
		return err
	}
	return nil
}


func (store *LinkRateHandler)init() error  {
	client,err:=utils.NewRedisClient(store.ip,store.port,6)
	if err!=nil{
		log.Fatalf("error connect redis instance %s:%d",store.ip,store.port)
	}
	log.Println("link rate handler set up redis client")
	store.redis=client

	go func() {
		log.Println("link rate handler worker started")
		if store.redis!=nil{
			defer store.redis.Close()
		}
		for{
			select {
			case <-store.done:
				log.Println("Link rate handler stop requested,exit")
				return
			case fn:=<-store.cache:
				lines,err:=utils.ReadLines(fn)
				if err!=nil{
					log.Println(err)
				}

				if len(lines)==0{
					log.Printf("%s empty file\n",fn)
					continue
				}

				ts,err:=utils.GetCreateTimeInSec(fn)
				if err!=nil{
					log.Println(err)
					continue
				}
				log.Printf("new file %s\n", fn)
				for _,line:=range lines{
					rate:=&models.Rate{
						Src:    "",
						Dst:    "",
						Volume: 0,
					}
					if err:=rate.Parse(line);err!=nil{
						log.Println(err)
						continue
					}
					if err:=store.StoreSingleRate(rate,ts);err!=nil{
						log.Println(err)
						continue
					}
				}

				if store.removeFile{
					err=utils.RMFile(fn)
					if nil!=err{
						log.Println(err)
					}
				}
			}
		}
	}()
	return nil
}

func (store *LinkRateHandler)destroy() error  {
	store.done<-common.StopSignal
	return nil
}


func NewLinkRateHandler(ip string,port int,rm bool) *LinkRateHandler  {
	return &LinkRateHandler{
		ip: ip,
		port: port,
		redis: nil,
		done:make(chan common.Signal,2),
		cache: make(chan string,102400),
		removeFile: rm,
	}
}

func (rh *LinkRateHandler)handle(fn string)  {
	log.Printf("new file %s\n",fn)
	rh.cache<-fn
}

func isRateFile(fn string)bool {
	return strings.Contains(fn,".rate")
}



