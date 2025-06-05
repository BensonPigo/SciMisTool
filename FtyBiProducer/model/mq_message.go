package model

// 用於封裝message
type DdlMessage struct {
	BatchID int64    `json:"BatchID"`
	XMLList []string `json:"XMLList"`
}

type DmlMessage struct {
	BatchID  int64    `json:"BatchID"`
	JSONList []string `json:"JSONList"`
}
