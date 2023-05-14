package main

import (
	"context"
	"log"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/chaosknight/biantrade/entity"
	"github.com/chaosknight/go-binance/v2/futures"
)

type Breed struct {
	InstId   string
	interval string
	kcandles []*entity.Candle
	tdcount  []int
	mutex    sync.RWMutex
}

func NewBreed(instId string, interval string) *Breed {
	return &Breed{
		InstId:   instId,
		interval: interval,
		kcandles: []*entity.Candle{},
		tdcount:  []int{},
	}
}

func (brd *Breed) LoadeK(client *futures.Client) bool {
	//时间判断是否真正需要加载数据
	brd.mutex.Lock()
	defer brd.mutex.Unlock()
	brd.kcandles = []*entity.Candle{}
	brd.tdcount = []int{}
	klines, err := client.NewKlinesService().Symbol(brd.InstId).Interval(brd.interval).Do(context.Background())
	klen := len(klines)
	if err == nil && klen > 0 {
		brd.tdcount = make([]int, klen)
		brd.kcandles = make([]*entity.Candle, klen)
		for i, v := range klines {
			o, _ := strconv.ParseFloat(v.Open, 64)
			c, _ := strconv.ParseFloat(v.Close, 64)
			h, _ := strconv.ParseFloat(v.High, 64)
			l, _ := strconv.ParseFloat(v.Low, 64)
			ts := time.Unix(v.OpenTime, 0)
			brd.kcandles[i] = &entity.Candle{
				O:  o,
				C:  c,
				H:  h,
				L:  l,
				TS: ts,
			}
		}

		SetTdcunt(brd.kcandles, brd.tdcount)
		log.Println("加载k线数据完成:", brd.InstId, len(brd.kcandles), "条")

	} else {
		log.Println(err.Error())
	}

	return len(brd.kcandles) > 0
}

func (brd *Breed) SetNewK(wsk *futures.WsKline) bool {
	brd.mutex.Lock()
	defer brd.mutex.Unlock()
	klen := len(brd.kcandles)
	if klen < 1 {
		return false
	}
	//删除过去太久数据，节省内存
	if klen > 1000 {

	}

	lastk := brd.kcandles[klen-1]

	o, _ := strconv.ParseFloat(wsk.Open, 64)
	c, _ := strconv.ParseFloat(wsk.Close, 64)
	h, _ := strconv.ParseFloat(wsk.High, 64)
	l, _ := strconv.ParseFloat(wsk.Low, 64)
	ts := time.Unix(wsk.StartTime, 0)
	k := &entity.Candle{
		O:  o,
		C:  c,
		H:  h,
		L:  l,
		TS: ts,
	}

	if k.TS == lastk.TS {
		lastk.O = k.O
		lastk.H = k.H
		lastk.L = k.L
		lastk.C = k.C
		lastk.Confirm = k.Confirm
	} else {
		brd.kcandles = append(brd.kcandles, k)
		brd.tdcount = append(brd.tdcount, 0)
	}
	SetTdcuntinde(brd.kcandles, brd.tdcount, len(brd.kcandles)-1)
	return true
}

func (brd *Breed) IsTiming() bool {
	brd.mutex.RLock()
	defer brd.mutex.RUnlock()
	size := len(brd.tdcount)
	abss := math.Abs(float64(brd.tdcount[size-1]))
	return abss >= 1
}

//是否已经加载k线数据
func (brd *Breed) IsEmpty() bool {
	brd.mutex.RLock()
	defer brd.mutex.RUnlock()
	return len(brd.kcandles) == 0
}

func (brd *Breed) GetTdc() (tdc int, ma7 float64, lastc float64) {
	brd.mutex.RLock()
	defer brd.mutex.RUnlock()
	tdc = 0
	ma7 = 0
	size := len(brd.tdcount)
	if size > 7 {
		tdc = brd.tdcount[size-1]
		for i := 1; i < 8; i++ {
			ma7 += brd.kcandles[size-i].C
		}
		ma7 = ma7 / 7
		lastc = brd.kcandles[size-1].C
	}
	return
}
