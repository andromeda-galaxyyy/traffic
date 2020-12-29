package main

import (
	"chandler.com/gogen/common"
	"chandler.com/gogen/utils"
	"encoding/json"
	"log"
	"sync"
)

type packager struct {
	flowDescs chan *flowDesc
	cache []*flowDesc
	threshold int
	classifierIP string
	classifierPort int
	doneChan chan common.Signal
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

func queryAndStore(ip string,port int,cache []*flowDesc)  {
	var wg sync.WaitGroup
	n:=len(cache)
	wg.Add(n)
	for idx,_:=range cache{
		desc :=cache[idx]
		go func(f *flowDesc) {
			defer wg.Done()
			report:=make(map[string]interface{})
			report["stats"]=[]float64{
				float64(f.MinPktSize),
				float64(f.MaxPktSize),
				f.MeanPktSize,
				f.DevPktSize,
				float64(f.MinInterval),
				float64(f.MaxInterval),
				f.MeanInterval,
				f.DevInterval,
			}
			req,err:=json.Marshal(report)
			if err!=nil{
				log.Println("error when marshal report")
				return
			}
			req=append(req,byte('*'))
			resp,err:=utils.SendAndRecv(ip,port,req,byte('*'))
			if err!=nil{
				log.Println("error when receive response")
				return
			}

			var obj classifierResp
			err=json.Unmarshal([]byte(resp),&obj)
			if err!=nil{
				log.Println("invalid response")
				return
			}

			desc.Pred=flowType(obj.resCode)
			//todo stats collect

		}(desc)
	}
	wg.Wait()
}


func (p *packager)Start()  {
	for{
		select {
		case <-p.doneChan:
			log.Println("Packager stop requested")
			return
		case f:=<-p.flowDescs:
			p.cache=append(p.cache,f)
			if len(p.cache)==p.threshold{
				tmp:=p.cache
				log.Println("cache is full,start query and store")
				go queryAndStore(p.classifierIP,p.classifierPort,tmp)
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




