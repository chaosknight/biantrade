package main

import (
	"github.com/chaosknight/biantrade/entity"
)

func SetTdcunt(kls []*entity.Candle, tdc []int) {
	if len(kls) != len(tdc) {
		return
	}
	for index, _ := range kls {
		SetTdcuntinde(kls, tdc, index)
	}
}

func SetTdcuntinde(kls []*entity.Candle, tdc []int, index int) {
	item := kls[index]
	if index < 4 || item.C == kls[index-4].C {
		tdc[index] = 0
	} else {
		if item.C > kls[index-4].C {
			if tdc[index-1] >= 0 {
				tdc[index] = tdc[index-1] + 1
			} else {
				tdc[index] = 1
			}
		} else {
			if tdc[index-1] <= 0 {
				tdc[index] = tdc[index-1] - 1
			} else {
				tdc[index] = -1
			}
		}
	}
}
