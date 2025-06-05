package model

import "encoding/xml"

type DmlMessage struct {
	BatchID  int      `json:"BatchID"`
	JSONList []string `json:"JSONList"`
}

// 用於封裝message
type DdlMessage struct {
	BatchID int      `json:"BatchID"`
	XMLList []string `json:"XMLList"`
}

// DdlData 對應 XML 結構
type DdlData struct {
	XMLName   xml.Name `xml:"DDLData"`
	EventData struct {
		Instance struct {
			TSQLCommand struct {
				CommandText string `xml:"CommandText"`
			} `xml:"TSQLCommand"`
		} `xml:"EVENT_INSTANCE"`
	} `xml:"EventData"`
}
