package entity

import (
	"fmt"
	"time"
)

type Candle struct {
	O           float64
	H           float64
	L           float64
	C           float64
	Vol         float64
	VolCcy      float64
	VolCcyQuote float64
	TS          time.Time
	Confirm     int
}

func (m *Candle) String() string {
	var str string
	str = fmt.Sprintf("%s\r\n┌------ Candle ------┐", str)
	if s := fmt.Sprintf("%v", m.O); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\nO:%v", str, m.O)
	}
	if s := fmt.Sprintf("%v", m.H); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\nH:%v", str, m.H)
	}
	if s := fmt.Sprintf("%v", m.L); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\nL:%v", str, m.L)
	}
	if s := fmt.Sprintf("%v", m.C); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\nC:%v", str, m.C)
	}
	if s := fmt.Sprintf("%v", m.Vol); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\nVol:%v", str, m.Vol)
	}
	if s := fmt.Sprintf("%v", m.VolCcy); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\nVolCcy:%v", str, m.VolCcy)
	}
	if s := fmt.Sprintf("%v", m.TS); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\nTS:%v", str, m.TS)
	}
	str = fmt.Sprintf("%s\r\n└----------- Candle ------------┘", str)
	return str
}
