package model

type Record struct {
	Sender               string `form:"sender" json:"sender"`
	Recevier             string `form:"recevier" json:"recevier"`
	SenderEncryptedCid   string `json:"secid"`
	RecevierEncryptedCid string `json:"recid"`
	Filename             string `json:"filename"`
	Message              string `json:"message"`
	Timestamp            string `json:"timestamp"`
}
