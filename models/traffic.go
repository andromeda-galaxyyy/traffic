package models

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Traffic struct {
	Src string
	Dst string
	Volume float64
}



func (r *Traffic)String() string  {
	return fmt.Sprintf("%s:%s:%.2f",r.Src,r.Dst,r.Volume)
}

func (r *Traffic)Parse(line string) error{
	if nil==r{
		log.Fatalf("nil pointer for rate struct")
	}
	elms:=strings.Split(line," ")
	if len(elms)!=3{
		return errors.New(fmt.Sprintf("error when parse %s to traffic sturct",line))
	}
	volume,err:=strconv.ParseFloat(elms[2],64)
	if err!=nil{
		return errors.New(fmt.Sprintf("error when parse %s to traffic struct %s",line,err))
	}
	r.Volume= volume
	r.Src=elms[0]
	r.Dst=elms[1]
	return nil
}
