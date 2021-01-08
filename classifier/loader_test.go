package main

import (
	"fmt"
	"log"
	"testing"
)

func TestLoader_Load(t *testing.T)  {
	var err error
	defaultLoader,err=newLoader("/home/stack/code/graduate/sim/traffic/pkts/video")
	if err!=nil{
		log.Fatalln(err)
	}
	err=defaultLoader.load(defaultLabeler)
	fn:=defaultLoader.randomPick()
	fmt.Println(fn)
}
