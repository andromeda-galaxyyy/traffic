package main

import (
	"chandler.com/gogen/common"
	"chandler.com/gogen/models"
	"chandler.com/gogen/utils"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	rport int=6379
	rip string="localhost"
)

func setUpFlowCounterReader(ip string,port int)error  {
	counterReader= models.NewDefaultCounterReader(ip,port)
	err:=counterReader.Init()
	return err

}


func setUpRedisHandle(ip string,port int) error  {
	ctx:=context.Background()
	delayHandle0 =redis.NewClient(&redis.Options{
		Addr:fmt.Sprintf("%s:%d",ip,port),
		Password: "",
		DB:0,
	})
	_,err:= delayHandle0.Ping(ctx).Result()
	if err!=nil{
		return err
	}

	delayHandle1 =redis.NewClient(&redis.Options{
		Addr:fmt.Sprintf("%s:%d",ip,port),
		Password: "",
		DB:1,
	})
	_,err= delayHandle1.Ping(ctx).Result()
	if err!=nil{
		return err
	}

	 lossHandle0=redis.NewClient(&redis.Options{
		Addr:fmt.Sprintf("%s:%d",ip,port),
		Password: "",
		DB:2,
	})
	_,err= lossHandle0.Ping(ctx).Result()
	if err!=nil{
		return err
	}

	lossHandle1=redis.NewClient(&redis.Options{
		Addr:fmt.Sprintf("%s:%d",ip,port),
		Password: "",
		DB:3,
	})
	_,err= lossHandle1.Ping(ctx).Result()
	if err!=nil{
		return err
	}
	return nil
}

func setUpLinkRateRedisHandle(ip string,port int,db int) error  {
	var err error
	linkRateRedisHandle,err=utils.NewRedisClient(ip,port,db)
	if err!=nil{
		return err
	}
	return nil
}

func setUpTelemetryRedisHandle(ip string,port int) error{
	telemetryHandle=redis.NewClient(&redis.Options{
		Addr:fmt.Sprintf("%s:%d", ip,port),
		DB:7,
	})
	_,err:=telemetryHandle.Ping(context.Background()).Result()
	if err!=nil{
		return err
	}
	return nil
}

func setUpTrafficMatrixHandler(ip string,port int)  error{
	var err error=nil
	trafficHandler,err=utils.NewRedisClient(ip,port,4)
	if err!=nil{
		return err
	}
	return nil
}

func loadTopo(fn string) error  {
	topo=make([][]bool,100)
	for i:=0;i<100;i++{
		topo[i]=make([]bool,100)
	}
	//load topo
	f,err:=ioutil.ReadFile(fn)
	if err!=nil{
		log.Fatalln(err)
	}
	log.Println("load done")
	obj:=make(map[string][][][]int)
	err=json.Unmarshal([]byte(f),&obj)
	if err!=nil{
		log.Fatalln(err)
	}
	//log.Println(obj)
	links,ok:=obj["topo"]
	if !ok{
		//log.Fatalf("invalid file")
		return errors.New("invalid topo file")
	}

	count:=0
	for i:=0;i<100;i++{
		for j:=0;j<100;j++{
			if links[i][j][0]!=-1{
				count++
				topo[i][j]=true
				topo[j][i]=true
			}
		}
	}
	if count!=566{
		return errors.New("invalid topo file,invalid topo")
	}
	return nil

}


func main()  {
	serverPort:=flag.Int("port",10086,"Server listening port")
	redisPort:=flag.Int("rport",6379,"Redis instance port")
	redisIp:=flag.String("rip","10.211.55.2","Redis instance ip")
	topoFn:=flag.String("topo","/home/stack/code/simulation/static/topo.json","Topo json")


	flag.Parse()
	if err:=loadTopo(*topoFn);err!=nil{
		log.Fatalf("error when loading topo file:%s\n",err)
	}
	rport=*redisPort
	rip=*redisIp
	//laod topo

	err:=setUpRedisHandle(rip,rport)
	if err!=nil{
		log.Fatalf("Cannot connect to redis instance %s:%d",rip,rport)
	}

	err=setUpFlowCounterReader(rip,rport)
	if err!=nil{
		log.Fatalf("Error init counter reader\n")
	}

	err=setUpTelemetryRedisHandle(rip, rport)
	if err!=nil{
		log.Fatalf("Error init network telemetry redis handle\n")
	}

	err=setUpLinkRateRedisHandle(rip,rport,6)
	if err!=nil{
		log.Fatalf("error init link rate redis handle %s\n",err)
	}

	err=setUpTrafficMatrixHandler(rip,rport)
	if err!=nil{
		log.Fatalf("error init traffic matrix handler %s\n",err)
	}

	sigs:=make(chan os.Signal,1)
	signal.Notify(sigs,syscall.SIGINT,syscall.SIGTERM,syscall.SIGKILL)

	quit:=make(chan common.Signal,1)
	go func() {
		<-sigs
		log.Println("Stop requested")
		log.Println("Send stop signal to server")
		quit<-common.StopSignal
	}()

	router=gin.Default()
	router.GET("/delay",GetDelayBetween)
	router.GET("/loss",GetLossBetween)
	router.GET("/flowcounter",GetFlowCounter)
	router.GET("/telemetry/loss",TelemetryFunc(false))
	router.GET("/telemetry/delay",TelemetryFunc(true))
	router.GET("/linkrate",getAllLinkRate)
	router.GET("/maxrate",getMaxRate)
	router.GET("/classifier",TimestampComplete(),getClassifierTestCase)
	router.GET("/traffic",getTrafficMatrix)

	server:=&http.Server{
		Addr: fmt.Sprintf(":%d",*serverPort),
		Handler: router,
	}
	go func() {
		if err:=server.ListenAndServe();err!=nil{
			log.Fatalln("Cannot start server")
		}
		log.Println("server started")
	}()

	<-quit
	log.Println("Server stop requested")
	ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()
	if err:=server.Shutdown(ctx);err!=nil{
		log.Fatalln("Cannot shutdown server",err)
	}
	log.Println("Server shutdown now")
}
