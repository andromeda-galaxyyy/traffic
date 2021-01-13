package models

import "encoding/json"

type TestStats struct {
	Ts            int64   `json:"ts"`
	NInstance     int     `json:"n_instance"`
	FalsePositive float64 `json:"false_positive"`
	FalseNegative float64 `json:"false_negative"`
	TruePositive  float64 `json:"true_positive"`
	TrueNegative  float64 `json:"true_negative"`

	Precision float64 `json:"precision"`

	PositivePredictValue float64 `json:"positive_predict_value"`
	FalseDiscoveryRate   float64 `json:"false_discovery_rate"`
	NegativePredictValue float64 `json:"negative_predict_value"`
	FalseOmissionRate    float64 `json:"false_omission_rate"`
}


func (s *TestStats) Box() ([]byte,error)  {
	return json.Marshal(*s)
}
func (s *TestStats) UnBox(data []byte) error  {
	return json.Unmarshal(data,s)
}