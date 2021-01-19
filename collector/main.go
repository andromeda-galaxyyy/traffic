package main

import (
	"chandler.com/gogen/utils"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	defaultDirs string
	rport int=6379
	rip string="localhost"
)


func main()  {
	base_dir :=flag.String("base","/tmp/listener_log","Base directory to watch")
	redisPort:=flag.Int("rport",6379,"Redis instance port")
	redisIp:=flag.String("rip","10.211.55.2","Redis instance ip")
	lossDelayDirs :=flag.String("lossDelayDirs","/tmp/rxloss,/tmp/rxdelay","Directory to watch")
	linkRateDirs:=flag.String("rate_dirs","/tmp/data","Directory to store link rate dirs")
	removeFile:=flag.Bool("rm",false,"Whether remove file")
	trafficMatrixDirs:=flag.String("traffic_dirs","/tmp/data","Directory to store traffic matrix")

	mode:=flag.Int("mode", 0, "Watcher function,0 for delay or loss,1 for link rate")

	flag.Parse()


	rport=*redisPort
	rip=*redisIp

	sigs:=make(chan os.Signal,1)
	signal.Notify(sigs,syscall.SIGINT,syscall.SIGTERM,syscall.SIGKILL)
	//dd:=make([]string,0)

	//id:=uint64(0)
	registration:=NewRegistration()

	handlers:=make([]Handler,0)
	preds:=make([]Pred,0)
	watchedDirectories:=make([][]string,0)

	//var handler Handler=nil
	//var pred Pred=nil

	enableLossDelayWatch:=false
	enableLinkRateWatch:=false
	enableTrafficMatrix:=false
	if *mode==-1{
		enableLinkRateWatch=true
		enableLossDelayWatch=true
	}else if *mode==0{
		enableLossDelayWatch=true
	}else if *mode==1{
		enableLinkRateWatch=true
		enableTrafficMatrix=true
	}

	if enableLossDelayWatch{
		delayOrLossDirectories:=make([]string,0)
		log.Println("delay or loss watcher")
		filepath.Walk(*base_dir, func(path string, info os.FileInfo, err error) error {
			if info != nil && info.IsDir() {
				delayOrLossDirectories = append(delayOrLossDirectories, path)
			}
			if err != nil {
				log.Fatalf("Error when scanning base directory %s\n", *base_dir)
			}
			return nil
		})

		for _, d := range strings.Split(*lossDelayDirs, ",") {
			if utils.IsDir(d) {
				delayOrLossDirectories = append(delayOrLossDirectories, d)
			}
		}

		handlers=append(handlers,NewLossDelayHandler(*redisIp, *redisPort, *removeFile))
		preds=append(preds,IsLossDelayFile)
		watchedDirectories=append(watchedDirectories,delayOrLossDirectories)
	}
	if enableLinkRateWatch{
		linkRateDirectories:=make([]string,0)
		log.Println("link rate watcher")
		preds=append(preds,isRateFile)
		handlers=append(handlers,NewLinkRateHandler(*redisIp, *redisPort, *removeFile))

		ds, _ := utils.SplitArgs(*linkRateDirs, ",")
		linkRateDirectories = append(linkRateDirectories, ds...)
		watchedDirectories=append(watchedDirectories,linkRateDirectories)
	}
	if enableTrafficMatrix{
		log.Println("enable traffic matrix file watcher")
		preds=append(preds,isTrafficMatrixFile)
		handlers=append(handlers,NewDefaultTrafficMatrixHandler(*redisIp,*redisPort,*removeFile))
		trafficDirs:=make([]string,0)
		ds,_:=utils.SplitArgs(*trafficMatrixDirs,",")
		trafficDirs=append(trafficDirs,ds...)
		watchedDirectories=append(watchedDirectories,trafficDirs)
	}

	log.Printf("handlers %d\n",len(handlers))
	for _,handler:=range handlers{
		err:=handler.init()
		if err!=nil{
			log.Fatalln(err)
		}
	}
	log.Println("all handler initiated")

	registrationIds:=make([]uint64,0)
	for idx,handler:=range handlers{
		id,err:=registration.Register(watchedDirectories[idx],handler,preds[idx])
		if err!=nil{
			log.Fatalln(err)
		}
		registrationIds=append(registrationIds,id)
	}
	log.Println("all handler registered")

	for _,id:=range registrationIds{
		err:=registration.Start(id)
		if err!=nil{
			log.Fatalln(err)
		}
	}

	log.Println("registration started")

	<-sigs
	log.Println("stop requested")
	for _,id:=range registrationIds{
		err:=registration.Stop(id)
		if err!=nil{
			log.Println(err)
		}
	}


	log.Println("registration stop")

	for _,h:=range handlers{
		if err:=h.destroy();err!=nil{
			log.Println(err)
		}
	}

	log.Println("Handler destroy")
	time.Sleep(2*time.Second)
	log.Println("All work done,exit")
}







//delayChan:=make(chan string,10240)
//lossChan:=make(chan string,10240)
//fileChan:=make(chan string,10240)
//doneChanToWatcher :=make(chan common.Signal,1)
//doneChanToWriter:=make(chan common.Signal,1)
//
//sigs:=make(chan os.Signal,1)
//signal.Notify(sigs,syscall.SIGINT,syscall.SIGTERM,syscall.SIGKILL)
//
//
//redisWriter := NewLossDelayHandler(*redisIp,*redisPort)
//redisWriter.delayChan=delayChan
//redisWriter.lossChan=lossChan
//redisWriter.fileChan=fileChan
//redisWriter.doneChan=doneChanToWriter
//
//err:=redisWriter.init()
//if err!=nil{
//	log.Fatalf("Error when connect to redis instance")
//}
//go redisWriter.Start()
//
//watcher:=&Watcher{
//	id:        0,
//	dirs:      dd,
//	done:      doneChanToWatcher,
//	worker:    nil,
//	delayChan: delayChan,
//	lossChan:  lossChan,
//	fileChan:  fileChan,
//	isDelay:   IsDelayFile,
//}
//watcher.init()
//go watcher.Start()
//
//quit:=make(chan common.Signal,1)
//go func() {
//	<-sigs
//	log.Println("Stop requested")
//	log.Println("Send stop signal to watcher")
//	doneChanToWatcher<-common.StopSignal
//	log.Println("Send stop signal to writer")
//	doneChanToWatcher<-common.StopSignal
//	log.Println("Send stop signal to server")
//	quit<-common.StopSignal
//}()

//<-quit
//log.Println("Watcher exits")
