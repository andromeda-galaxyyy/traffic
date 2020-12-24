package models

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Rate struct {
	Src string
	Dst string
	Volume uint64
}

func (r *Rate)String() string  {
	return fmt.Sprintf("%s:%s:%d",r.Src,r.Dst,r.Volume)
}

func (r *Rate)Parse(line string) error{
	if nil==r{
		log.Fatalf("nil pointer for rate struct")
	}
	elms:=strings.Split(line,":")
	if len(elms)!=3{
		return errors.New(fmt.Sprintf("error when parse %s to rate sturct",line))
	}
	volume,err:=strconv.ParseInt(elms[2],10,64)
	if err!=nil{
		return errors.New(fmt.Sprintf("error when parse %s to rate struct %s",line,err))
	}
	r.Volume= uint64(volume)
	r.Src=elms[0]
	r.Dst=elms[1]
	return nil
}