package main

import (
	"encoding/json"
)

type flowType int

const (
	video flowType=iota
	iot flowType=iota
	voip flowType=iota
	ar flowType=iota
	unknown flowType=iota
)

type result struct {
	Ts         int64  `json:"ts"`
	Fn         string `json:"fn"`
	FlowId     int64  `json:"flow_id"`
	Label      string `json:"label"`
	Prediction string `json:"prediction"`
	IsValid    bool   `json:"is_valid"`
}



func (r *result) box()([]byte,error){
	return json.Marshal(*r)
}

func (r *result)unbox(data []byte) error{
	return json.Unmarshal(data,r)
}


type flowDesc struct {
	FileName string
	//pkt size
	MaxPktSize float64
	MinPktSize float64
	MeanPktSize float64
	DevPktSize float64

	//pkt interval
	MaxInterval  float64
	MinInterval  float64
	MeanInterval float64
	DevInterval float64

	//label
	TrueLabel flowType
	Pred flowType
}

