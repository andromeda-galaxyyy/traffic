package main

import (
	"chandler.com/gogen/common"
	"chandler.com/gogen/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
)

type packager struct {
	flowDescs chan *flowDesc
	cache []*flowDesc
	threshold int
	classifierIP string
	classifierPort int
	doneChan chan common.Signal
	redisHandle *redis.Client
}

func (p *packager)setupRedis(ip string,port int,db int){
	var err error
	p.redisHandle,err=utils.NewRedisClient(ip,port,db)
	if err!=nil{
		log.Fatalln(err)
	}
}

func NewPackager(ip string,port int) *packager {
	return &packager{
		flowDescs:      make(chan *flowDesc,10240),
		cache:          make([]*flowDesc,0),
		threshold:      50,
		classifierIP:   ip,
		classifierPort: port,
		doneChan:       make(chan common.Signal,1),
	}
}

type classifierResp struct {
	resCode int `json:"res"`
}

func (p *packager)queryAndStore(ip string,port int,cache []*flowDesc)  {
	//var wg sync.WaitGroup
	//n:=len(cache)
	//wg.Add(n)
	req:=make(map[string]interface{})
	req["num"]=len(cache)
	data:=make([]map[string][]float64,0)
	for _,f:=range cache{
		report:=make(map[string][]float64)
		report["stats"]=[]float64{
			f.MinPktSize,
			f.MaxPktSize,
			f.MeanPktSize,
			f.DevPktSize,
			f.MinInterval,
			f.MaxInterval,
			f.MeanInterval,
			f.DevInterval,
		}
		data=append(data,report)
	}

		req["data"]=data
		reqStr,err:=json.Marshal(req)
		if err!=nil{
			log.Println("error when marshal report")
			return
		}
		reqStr=append(reqStr,byte('*'))
		respStr,err:=utils.SendAndRecv(ip,port,reqStr,byte('*'))
		if err!=nil{
			log.Println("error when receive response")
			log.Println(err)
			return
		}
		resp:=make(map[string]interface{})
		err=json.Unmarshal([]byte(respStr),&resp)
		if err!=nil{
			log.Printf("Error parsing response %s\n",respStr)
		}
		/**
		{
			"num": 50,
			"res": [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]
		}
		 */
		labels,ok:=resp["res"].([]interface{})
		if !ok{
			log.Printf("Error parsing response %s\n",respStr)
		}

		for idx,desc:=range cache{
			desc.Pred=flowType(int(labels[idx].(float64)))
		}

		log.Println(resp)
		//log.Println(respStr)
		// 分析存储
		//  分为true negative 和false negative
		// https://en.wikipedia.org/wiki/Binary_classification
		n_tp:=0
		n_fp:=0
		n_fn:=0
		n_tn:=0

		for _,desc:=range cache{
			if desc.TrueLabel==video{
				if desc.Pred==video{
					n_tp+=1
				}else{
					n_fn+=1
				}
			}else{
				if desc.Pred==video{
					n_fp+=1
				}else{
					n_tn+=1
				}
			}
		}

		fp:=0.0
		if n_tn+n_fp!=0.0{
			fp=float64(n_fp)/float64(n_tn+n_fp)
		}
		fn:=0.0
		if n_tp+n_fn!=0.0{
			fn=float64(n_fn)/float64(n_tp+n_fn)
		}
		tp:=0.0
		if n_tp+n_fn!=0.0{
			tp=float64(n_tp)/float64(n_tp+n_fn)
		}
		tn:=0.0
		if n_tn+n_fp!=0.0{
			tn=float64(n_tn)/float64(n_tn+n_fp)
		}

		test:=&testStats{
			Ts:            utils.NowInMilli(),
			NInstance:     len(cache),
			FalsePositive: fp,
			FalseNegative: fn,
			TruePositive:  tp,
			TrueNegative:  tn,
		}
		//todo write to redis
		testStr,err:=test.marshal()
		if err!=nil{
			log.Printf("unable to marshal test case %s\n",test)
		}
		err=p.redisHandle.ZAdd(context.Background(),fmt.Sprintf("%d",test.Ts),&redis.Z{
			Score: float64(test.Ts),
			Member: testStr,
		}).Err()
		if err!=nil{
			log.Println("unable to store in redis")
		}


}




func (p *packager)Start()  {
	log.Println("package start")
	for{
		select {
		case <-p.doneChan:
			log.Println("Packager stop requested")
			return
		case f:=<-p.flowDescs:
			//log.Println("accept")
			p.cache=append(p.cache,f)
			if len(p.cache)==p.threshold{
				tmp:=p.cache
				log.Println("cache is full,start query and store")
				//log.Println(tmp)
				go p.queryAndStore(p.classifierIP,p.classifierPort,tmp)
				p.cache=make([]*flowDesc,0)
			}
		}
	}
}

func (p *packager)Stop()  {
	go func() {
		p.doneChan<-common.StopSignal
	}()
}




