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

type trafficMatrixHandler struct {
	ip string
	port int
	redis *redis.Client
	done chan common.Signal
	cache chan string
	removeFile bool
}

func (t *trafficMatrixHandler)storeTrafficMatrix(traffics []*models.Traffic,score int64)error{
	ctx:=context.Background()
	pipe:=t.redis.Pipeline()
	for _,traffic:=range traffics{
		key:=fmt.Sprintf("%s-%s",traffic.Src,traffic.Dst)
		//content:=traffic.String()
		if err:=t.redis.ZAdd(ctx,key,&redis.Z{
			Score: float64(score),
			Member: traffic.Volume,
		}).Err();err!=nil{
			log.Println(err)
			return err
		}
	}
	_,err:=pipe.Exec(ctx)
	if err!=nil{
		return err
	}
	//for _,cmder:=range cmders{
	//	cmd,ok:=cmder.(*redis.StringCmd)
	//	if !ok{
	//		log.Fatalln(errors.New("error cast cmd to string cmd"))
	//	}
	//	if cmd.Err()!=nil{
	//		return err
	//	}
	//}
	return nil
}

func (h *trafficMatrixHandler)init() error{
	client,err:=utils.NewRedisClient(h.ip,h.port,4)
	if err!=nil{
		log.Fatalf("traffic matrix set up redis client error with ip:%s,port:%d",h.ip,h.port)
	}
	log.Println("traffic matrix set up redis client success")
	h.redis=client
	go func() {
		log.Println("traffic matrix handler worker started")
		if h.redis!=nil{
			defer h.redis.Close()
		}
		for {
			select {
			case <-h.done:
				log.Println("traffic matrix handler stop requested")
				return
			case fn:=<-h.cache:
				lines,err:=utils.ReadLines(fn)
				if err!=nil{
					log.Printf("error when traffic matrix handler read file :%s\n",fn)
					continue
				}
				if len(lines)==0{
					log.Printf("emptry file:%s\n",fn)
					continue
				}
				ts,err:=utils.GetCreateTimeInSec(fn)
				//log.Println(ts)
				if err!=nil{
					log.Printf("error when traffic matrix handler get create time of %s\n",fn)
					continue
				}
				traffics:=make([]*models.Traffic,0)
				for _,line:=range lines{
					//empty line
					if len(line)==0{
						continue
					}

					traffic:=&models.Traffic{
						Src:    "",
						Dst:    "",
						Volume: 0,
					}
					if err:=traffic.Parse(line);err!=nil{
						log.Println(err)
						continue
					}

					traffics=append(traffics,traffic)
				}
				if err:=h.storeTrafficMatrix(traffics,ts);err!=nil{
					log.Printf("error when traffic matrix handler store traffics %s\n",fn)
					continue
				}


				if h.removeFile{
					if err:=utils.RMFile(fn);err!=nil{
						log.Printf("error when traffic matrix handler remove file %s\n",fn)
					}
				}
			}
		}
	}()
	return nil
}

func (h *trafficMatrixHandler)destroy() error  {
	h.done<-common.StopSignal
	return nil
}

func (h *trafficMatrixHandler)handle(fn string)  {
	log.Printf("traffic matrix handler new file send to cache:%s\n",fn)
	h.cache<-fn
}

func isTrafficMatrixFile(fn string) bool  {
	return strings.Contains(fn,".matrix")
}

func NewDefaultTrafficMatrixHandler(ip string,port int,removeFile bool) *trafficMatrixHandler{
	return &trafficMatrixHandler{
		ip:         ip,
		port:       port,
		redis:      nil,
		done:       make(chan common.Signal,2),
		cache:      make(chan string,102400),
		removeFile: removeFile,
	}
}