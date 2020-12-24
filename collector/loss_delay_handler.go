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



type LossDelayHandler struct {
	ip        string
	port      int
	//handle0 and handle1 is for delay,src && dst respectively
	handle0   *redis.Client
	handle1 *redis.Client
	//handle2 and handle3 is for loss,src && dst respectively
	handle2 *redis.Client
	handle3 *redis.Client

	lossChan  chan string
	delayChan chan string
	fileChan  chan string
	doneChan  chan common.Signal

	removeFile bool
	// redis采用多副本，以src为key和以dst为key
	// 方便debug和查询

}

func NewLossDelayHandler(ip string,port int,rm bool) *LossDelayHandler {
	return &LossDelayHandler{
		ip:        ip,
		port:      port,
		lossChan:  make(chan string,102400),
		delayChan: make(chan string,102400),
		fileChan:  make(chan string,102400),
		doneChan:  make(chan common.Signal),
		removeFile: rm,
	}
}


func (ld *LossDelayHandler)handle(fn string)  {
	if strings.Contains(fn,".loss"){
		ld.lossChan<-fn
	}else{
		ld.delayChan<-fn
	}
}

func IsLossDelayFile(fn string) bool {
	return strings.Contains(fn,".loss")||strings.Contains(fn,".delay")
}

func (ld *LossDelayHandler)destroy() error{
	ld.doneChan<-common.StopSignal
	return nil
}



func (r *LossDelayHandler) BatchWriteLossStats(lines []string, ts int64) error {
	for _,line:=range lines{
		_=r.WriteLossStats(line,ts)
	}
	return nil
}


func (r *LossDelayHandler) WriteDelayStats(line string, ts int64) error {
	desc:=&models.FlowDesc{}
	err:= models.DescFromDelayStats(desc,line)
	if err!=nil{
		return err
	}
	srcIp:=desc.SrcIP
	srcId,err:=utils.IdFromIP(srcIp)
	if err!=nil{
		return err
	}
	dstIp:=desc.DstIP
	dstId,err:=utils.IdFromIP(dstIp)
	if err!=nil{
		return err
	}
	ctx:=context.Background()
	if err:=r.handle0.ZAdd(ctx,fmt.Sprintf("%d",srcId),&redis.Z{
		Score:  float64(ts),
		Member: line,
	}).Err();err!=nil{
		return err
	}

	if err:=r.handle1.ZAdd(ctx,fmt.Sprintf("%d",dstId),&redis.Z{
		Score:  float64(ts),
		Member: line,
	}).Err();err!=nil{
		return err
	}
	return nil
}



func (r *LossDelayHandler) WriteLossStats(line string, ts int64) error {
	desc:=&models.FlowDesc{}
	err:= models.DescFromRxLossStats(desc,line)
	if err!=nil{
		return err
	}
	srcIp:=desc.SrcIP
	srcId,err:=utils.IdFromIP(srcIp)
	if err!=nil{
		return err
	}
	dstIp:=desc.DstIP
	dstId,err:=utils.IdFromIP(dstIp)
	if err!=nil{
		return err
	}
	ctx:=context.Background()
	if err:=r.handle2.ZAdd(ctx,fmt.Sprintf("%d",srcId),&redis.Z{
		Score:  float64(ts),
		Member: line,
	}).Err();err!=nil{
		return err
	}

	if err:=r.handle3.ZAdd(ctx,fmt.Sprintf("%d",dstId),&redis.Z{
		Score:  float64(ts),
		Member: line,
	}).Err();err!=nil{
		return err
	}
	return nil
}





func (r *LossDelayHandler) BatchWriteDelayStats(lines []string, ts int64) error {
	for _,line:=range lines{
		_=r.WriteDelayStats(line,ts)
	}
	return nil
}


func (r *LossDelayHandler) init() error {
	r.handle0 =redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",r.ip, r.port),
		Password: "",
		DB: 0,
	})
	ctx:=context.Background()
	_,err:=r.handle0.Ping(ctx).Result()
	if err!=nil{
		return err
	}

	r.handle1 =redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",r.ip, r.port),
		Password: "",
		DB: 1,
	})
	//ctx:=context.Background()
	_,err=r.handle1.Ping(ctx).Result()
	if err!=nil{
		return err
	}

	r.handle2 =redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",r.ip, r.port),
		Password: "",
		DB: 2,
	})
	_,err=r.handle2.Ping(ctx).Result()
	if err!=nil{
		return err
	}

	r.handle3 =redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",r.ip, r.port),
		Password: "",
		DB: 3,
	})
	_,err=r.handle3.Ping(ctx).Result()
	if err!=nil{
		return err
	}

	go func() {
		log.Println("loss delay handler worker started")
		if nil!=r.handle0 {
			defer r.handle0.Close()
		}
		//ctx:=context.Background()
		for {
			select {
			case <-r.doneChan:
				log.Println("Redis writer stop requested")
				return
			case  fn:=<-r.lossChan:
				if !utils.IsFile(fn) {
					continue
				}

				log.Printf("loss file:%s\n", fn)
				lines, err := utils.ReadLines(fn)

				lines=lines[1:]

				if err != nil {
					log.Printf("Redis writer cannot read line from %s\n", fn)
					continue
				}
				ctime,_:=utils.GetCreateTimeInSec(fn)
				err=r.BatchWriteLossStats(lines, ctime)
				if err != nil {
					log.Printf("Redis writer cannot store %s\n", fn)
					continue
				}
				if r.removeFile{
					err=utils.RMFile(fn)
					if err!=nil{
						log.Println(err)
					}
				}
				continue
			case fn := <-r.delayChan:
				if !utils.IsFile(fn) {
					continue
				}

				log.Printf("delay file:%s\n", fn)
				lines, err := utils.ReadLines(fn)

				lines=lines[1:]

				if err != nil {
					log.Printf("Redis writer cannot read line from %s\n", fn)
					continue
				}
				ctime,_:=utils.GetCreateTimeInSec(fn)
				err=r.BatchWriteDelayStats(lines, ctime)

				if err != nil {
					log.Printf("Redis writer cannot store %s\n", fn)
					continue
				}
				if r.removeFile{
					err=utils.RMFile(fn)
					if err!=nil{
						log.Println(err)
					}
				}
			}
		}
	}()
	return nil
}




