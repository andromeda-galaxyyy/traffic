package main

import (
	"chandler.com/gogen/utils"
	"flag"
	"fmt"
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
	dirs:=flag.String("dirs","/tmp/rxloss,/tmp/rxdelay","Directory to watch")
	removeFile:=flag.Bool("rm",false,"Whether remove file")

	mode:=flag.Int("mode", 0, "Watcher function,0 for delay or loss,1 for link rate")

	flag.Parse()


	rport=*redisPort
	rip=*redisIp

	sigs:=make(chan os.Signal,1)
	signal.Notify(sigs,syscall.SIGINT,syscall.SIGTERM,syscall.SIGKILL)
	dd:=make([]string,0)

	id:=uint64(0)
	registration:=NewRegistration()

	var handler Handler=nil
	var pred Pred=nil


	if *mode==0{
		fmt.Println("delay or loss watcher")
		filepath.Walk(*base_dir, func(path string, info os.FileInfo, err error) error {
			if info!=nil&&info.IsDir(){
				dd=append(dd,path)
			}
			if err!=nil{
				log.Fatalf("Error when scanning base directory %s\n",*base_dir)
			}
			return nil
		})

		for _,d:=range strings.Split(*dirs,","){
			if utils.IsDir(d){
				dd=append(dd,d)
			}
		}

		handler= NewLossDelayHandler(*redisIp,*redisPort,*removeFile)
		pred=IsLossDelayFile
	}else if *mode==1{
		log.Println("link rate watcher")
		pred=isRateFile
		handler=NewLinkRateHandler(*redisIp,*redisPort,*removeFile)
		ds,_:=utils.SplitArgs(*dirs,",")
		dd=append(dd,ds...)
	}

	log.Printf("watch dirs %s\n",dd)

	err:= handler.init()
	if err!=nil{
		log.Fatalln(err)
	}
	log.Println("handler started")
	id,err=registration.Register(dd, handler,pred)
	if err!=nil{
		log.Fatalln(err)
	}
	err=registration.Start(id)
	if err!=nil{
		log.Fatalln(err)
	}
	log.Println("registration started")

	<-sigs
	log.Println("stop requested")
	err=registration.Stop(id)
	if err!=nil{
		log.Println(err)
	}
	log.Println("registration stop")
	if err:=handler.destroy();err!=nil{
		log.Println(err)
	}
	log.Println("Handler destroy")
	time.Sleep(2*time.Second)
	log.Println("All work done,exit")






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
}
