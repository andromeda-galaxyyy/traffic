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
	"time"
)

var (
	defaultLoader *loader
)


type loader struct {
	dir             string
	labels          map[string]flowType
	files           []string
	labeledFiles map[flowType][]string
	duplicateFactor int
}


func newLoader(dir string) (l *loader,err error){
	if !utils.IsDir(dir){
		return nil,errors.New(fmt.Sprintf("dir %s not exits",dir))
	}
	l=&loader{
		dir:    dir,
		labels: make(map[string]flowType),
		files: make([]string,0),
		duplicateFactor: 300,
	}
	l.labeledFiles=make(map[flowType][]string)
	l.labeledFiles[video]=make([]string,0)
	l.labeledFiles[iot]=make([]string,0)
	l.labeledFiles[voip]=make([]string,0)
	l.labeledFiles[ar]=make([]string,0)
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

			label:=ll(path)
			l.labels[path]=label
			l.labeledFiles[label]=append(l.labeledFiles[label],path)
			l.files=append(l.files,path)

		}
		return nil
	})
	videoFns:=l.labeledFiles[video]
	if len(videoFns)==0{
		log.Println("warn: find no video pkts")
	}
	for i:=0;i<l.duplicateFactor;i++{
		l.files=append(l.files,videoFns...)
	}
	//shuffle
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(l.files), func(i, j int) {
		l.files[i], l.files[j] = l.files[j], l.files[i]
	})
	log.Println("shuffle done")
	return nil

}

func (l *loader)getFlowType(fn string) (res flowType){
	if res,ok:=l.labels[fn];ok{
		return res
	}
	return unknown
}


