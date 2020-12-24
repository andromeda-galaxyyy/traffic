package main

import (
	"chandler.com/gogen/common"
	"chandler.com/gogen/utils"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
)

var dirNotExistsErr=errors.New("directory not exists")

type Handler interface {
	handle(fn string)
	init() error
	destroy() error
}

type Pred func(string) bool

type Registration struct {
	count uint64
	data map[uint64]*watcher
}


type watcher struct {
	started bool
	id uint64
	dirs []string
	done chan common.Signal
	handle *fsnotify.Watcher
	handler Handler
	pred Pred
}

func NewRegistration() *Registration  {
	return &Registration{
		data: make(map[uint64]*watcher),
		count: 1,
	}
}

func (r *Registration)Register(dirs []string,handler Handler,pred Pred) (uint64,error){
	if 0==len(dirs){
		return 0,errors.New("length of dirs is 0")
	}
	for _,dir:=range dirs{
		if !utils.IsDir(dir){
			return 0,errors.New(fmt.Sprintf("dir %s not exists",dir))
		}
	}
	w,err:=fsnotify.NewWatcher()
	if err!=nil{
		return 0,err
	}

	r.data[r.count]=&watcher{
		started: false,
		id:      r.count,
		dirs:    dirs,
		done:    make(chan common.Signal,1),
		handle:  w,
		handler: handler,
		pred: pred,
	}
	watcher:=r.data[r.count]
	for _,dir:=range dirs{
		err=watcher.handle.Add(dir)
		if err!=nil{
			log.Fatalf("error watch dir %s\n",dir)
			return 0,err
		}
	}
	res:=r.count
	r.count+=1
	return res,nil

}

func (r *Registration)Stop(id uint64) error{
	if watcher,ok:=r.data[id];ok{
		watcher.done<-common.StopSignal
		return nil
	}
	return errors.New(fmt.Sprintf("invalid id %d",id))
}

func (r *Registration)Start(id uint64) error {
	if watcher,ok:=r.data[id];ok{
		if watcher.started{
			return errors.New(fmt.Sprintf("watcher %d already started", id))
		}
		go func() {
			watcher.started=true
			for{
				select {
				case event,ok:=<-watcher.handle.Events:
					if !ok{
						log.Fatalf("%d watcher channel pannic\n",id)
					}
					////log.Println("event:",event)
					//if event.Op&fsnotify.Create==fsnotify.Create{
					//	//log.Printf("[%s] create file :%s\n",w,event.Name)
					//}
					// 创建了新的文件
					if event.Op&fsnotify.Write==fsnotify.Write{
						//log.Printf("[%s] modified file: %s\n",w,event.Name)
						if utils.IsDir(event.Name){
							err:=watcher.handle.Add(event.Name)
							if nil!=err{
								log.Printf("cannot watch %s\n",event.Name)
							}
						}else if utils.IsFile(event.Name){
							if watcher.pred(event.Name){
								watcher.handler.handle(event.Name)
							}
						}
					}
				case <-watcher.done:
					log.Printf("Shutting done watcher %d\n",watcher.id)
					return
				}
			}
			//defer func() {watcher.started=false}()
			defer func() {
				watcher.started=false
				watcher.handle.Close()
			}()
		}()
	}
	return nil

}

func (r *Registration)StartAll() []error{
	errs:=make([]error,0)
	for id,_:=range r.data{
		err:=r.Start(id)
		if err!=nil{
			errs=append(errs,err)
		}
	}
	return errs
}

func (r *Registration)StopAll() []error{
	errs:=make([]error,0)
	for id,_:=range r.data{
		err:=r.Stop(id)
		if err!=nil{
			errs=append(errs,err)
		}
	}
	return errs
}