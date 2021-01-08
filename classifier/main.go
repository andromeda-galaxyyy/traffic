package main

import (
	"flag"
	"log"
)

func main()  {
	pktsDir:=flag.String("pkts_dir","/home/stack/code/graduate/sim/traffic/pkts/default","pkts dir")
	nWorker:=flag.Int("workers",1,"Number of workers")
	classifierIP:=flag.String("clsip","10.211.55.2","IP of classifier")
	classifierPort:=flag.Int("clsport",5000,"classifier port")
	interfaceName:=flag.String("intf","h0-eth0","interface to send packet")
	payloadSize:=flag.Int("pkt_size",1300,"Payload size")
	redisIP:=flag.String("rip","10.211.55.2","redis ip")
	redisPort:=flag.Int("rport",6379,"redis port")
	rdb:=flag.Int("rdb",8,"redis db")


	selfID:=flag.Int("id",0,"selfID")
	targetID:=flag.Int("target",1,"targetID")
	flag.Parse()

	log.Printf("selfid:%d\n",*selfID)
	log.Printf("targeid:%d\n",*targetID)

	c:=&controller{
		nWorkers: *nWorker,
		pktDir: *pktsDir,
		intf: *interfaceName,
		selfId: *selfID,
		targetId: *targetID,
		classifierIp: *classifierIP,
		classifierPort: *classifierPort,
		payloadPktSize: *payloadSize,
		redisPort: *redisPort,
		redisIp:*redisIP,
		redisDB: *rdb,
	}
	c.init()
	c.start()
}
