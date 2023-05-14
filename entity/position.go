package entity

import (
	"fmt"
)

type Position struct {
	Symbol   string
	Size     float64 //持仓数量
	Side     string  //持仓方向
	Price    float64 //持仓成本价
	NowPrice float64 //最新价格
	Persont  float64 //盈亏百分百
	UnPnl    float64 //盈亏
}

func (m *Position) SetNewPrice(newprice float64) {
	m.NowPrice = newprice
	m.UnPnl = (m.NowPrice - m.Price) * m.Size

	m.Persont = (m.NowPrice - m.Price) * 100 / m.Price
	if m.Side == "SHORT" {
		m.Persont = -1 * m.Persont
	}

}

func (m *Position) String() string {
	var str string
	str = fmt.Sprintf("\r\n%s┌------ Position ------┐", str)
	if s := fmt.Sprintf("%v", m.Symbol); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\nSymbol: %v", str, m.Symbol)
	}
	if s := fmt.Sprintf("%v", m.Size); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\n持仓数量: %v", str, m.Size)
	}

	if s := fmt.Sprintf("%v", m.Side); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\n持仓方向: %v", str, m.Side)
	}

	if s := fmt.Sprintf("%v", m.Price); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\n成交均价: %v", str, m.Price)
	}

	if s := fmt.Sprintf("%v", m.NowPrice); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\n最新价: %v", str, m.NowPrice)
	}

	if s := fmt.Sprintf("%v", m.UnPnl); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\n盈亏: %v", str, m.UnPnl)
	}

	if s := fmt.Sprintf("%v", m.Persont); s != "" && s != "0" {
		str = fmt.Sprintf("%s\r\n盈亏率: %v", str, m.Persont)
	}

	str = fmt.Sprintf("%s\r\n└----------- Position ------------┘", str)
	return str

}
