package main

import (
	"chandler.com/gogen/utils"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

var (
	defaultLoader *loader
)


type loader struct {
	dir             string
	labels          map[string]flowType
	files           []string
	duplicateFactor float64
}


func newLoader(dir string) (l *loader,err error){
	if !utils.IsDir(dir){
		return nil,errors.New(fmt.Sprintf("dir %s not exits",dir))
	}
	l=&loader{
		dir:    dir,
		labels: make(map[string]flowType),
		files: make([]string,0),
		duplicateFactor: 3,
	}
	return l,nil
}

func (l *loader)randomPick() (fn string)  {
	return l.files[rand.Intn(len(l.files))]
}


type labeler func(fn string) flowType

func defaultLabeler(fn string) flowType  {
	if strings.Contains(fn,"video"){
		return video
	}
	if strings.Contains(fn,"iot"){
		return iot
	}
	if strings.Contains(fn,"voip"){
		return voip
	}
	if strings.Contains(fn,"ar"){
		return ar
	}
	return unknown
}


func (l *loader)load(ll labeler) error{
	filepath.Walk(l.dir, func(path string, info os.FileInfo, err error) error {
		if info!=nil&&!info.IsDir(){
			//we meet a pkts file
			if !strings.Contains(info.Name(),".pkts"){
				return nil
			}
			log.Printf("meet a pkts file %s\n",path)
			l.labels[path]=ll(path)
			l.files=append(l.files,path)
		}
		return nil
	})
	return nil
}

func (l *loader)getFlowType(fn string) (res flowType){
	if res,ok:=l.labels[fn];ok{
		return res
	}
	return unknown
}


