package main

import "encoding/json"

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



type stats struct {
	Ts int64 `json:"ts"`
	NInstance int64 `json:"n_instance"`
	FalsePositive float64 `json:"false_positive"`
	FalseNegative float64 `json:"false_negative"`
	TruePositive float64 `json:"true_positive"`
	TrueNegative float64 `json:"true_negative"`

}

func (s *stats)marshal() ([]byte,error)  {
	return json.Marshal(*s)
}
func (s *stats)unbox(data []byte) error  {
	return json.Unmarshal(data,s)
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

