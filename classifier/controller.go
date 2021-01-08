package main

import (
	"chandler.com/gogen/common"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type controller struct {
	pktDir string
	wg *sync.WaitGroup
	nWorkers int
	workers []*worker
	packager *packager
	classifierIp string
	classifierPort int
	redisIp string
	redisPort int
	redisDB int

	intf string
	selfId int
	targetId int
	payloadPktSize int
}



func (c *controller)init()  {
	var err error
	defaultLoader,err=newLoader(c.pktDir)
	if err!=nil{
		log.Fatal(err)
	}
	err=defaultLoader.load(defaultLabeler)
	if err!=nil{
		log.Fatal(err)
	}
	c.packager=NewPackager(c.classifierIp,c.classifierPort)
	c.packager.setupRedis(c.redisIp,c.redisPort,c.redisDB)


	c.wg=&sync.WaitGroup{}
	c.wg.Add(c.nWorkers)
	c.workers=make([]*worker,0)
	for i:=0;i<c.nWorkers;i++{
		c.workers=append(c.workers,&worker{
			id:              i,
			doneChan:        make(chan common.Signal,1),
			wg: c.wg,
			flowDescs: c.packager.flowDescs,

			intf: c.intf,
			selfId: c.selfId,
			targetId: c.targetId,
			payloadPerPacketSize: c.payloadPktSize,
		})
	}
	for i:=0;i<c.nWorkers;i++{
		c.workers[i].Init()
	}
	sigs:=make(chan os.Signal,1)
	signal.Notify(sigs,syscall.SIGINT,syscall.SIGTERM,syscall.SIGKILL)
	go func() {
		sig:=<-sigs
		log.Printf("recevied signal %s,shuting down worker\n",sig)
		for wid,w:=range c.workers{
			log.Printf("Prepare to shutting done worker: %d\n",wid)
			w.doneChan<-common.StopSignal
		}
	}()

}

func (c *controller)start()  {
	go c.packager.Start()
	for _,w:=range c.workers{
		w.reset()
		go w.start()
	}
	c.wg.Wait()
	c.packager.Stop()
	log.Println("Controller exit")
}




