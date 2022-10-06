package core

import (
	"encoding/json"
	"time"
)

type Proof struct {
	Amount       int64  `json:"amount"`
	Secret       string `json:"secret" gorm:"primaryKey"`
	C            string `json:"C"`
	reserved     bool
	sendId       string
	timeCreated  time.Time
	timeReserved time.Time
}

type Proofs []Proof

type Promise struct {
	B_b    string `json:"C_b" gorm:"primaryKey"`
	C_c    string `json:"C_c"`
	Amount int64  `json:"amount"`
}

type BlindedMessages []BlindedMessage

type BlindedMessage struct {
	Amount int64  `json:"amount"`
	B_     string `json:"B_"`
}
type BlindedSignature struct {
	Amount int64  `json:"amount"`
	C_     string `json:"C_"`
}

func ToJson(i interface{}) string {
	b, err := json.Marshal(i)
	if err != nil {
		return err.Error()
	}
	return string(b)
}
