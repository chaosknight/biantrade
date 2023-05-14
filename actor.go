package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/chaosknight/biantrade/entity"
	"github.com/chaosknight/go-binance/v2/futures"
	"github.com/chaosknight/skynet/actor"
	"github.com/chaosknight/skynet/types"
)

const LoadBreed = "LoadBreed"

const OrderActor = "OrderActor"

func (sys *Sys) actorinit() {
	//加载历史k线数据
	breedactor := actor.NewFromReducer(LoadBreed, 20, func(a types.Actor, msg *types.MasterMsg) {
		log.Println("加载k线数据 :", msg.Cmd)
		cell, ok := sys.CellMap.Load(msg.Cmd)
		if ok {
			v := cell.(*Cell)
			if v.brd.IsEmpty() {
				v.brd.LoadeK(sys.fClient)
				time.Sleep(time.Duration(3) * time.Second)
			}
		}
	})
	sys.net.Rigist(breedactor, 1)

	//平仓
	emptyactor := actor.NewFromReducer(OrderActor, 1024, func(a types.Actor, msg *types.MasterMsg) {
		log.Println("策略执行中 :", msg.Cmd)
		instId := msg.Cmd
		sys.pmux.Lock()
		defer sys.pmux.Unlock()

		//http请求，positon数据还未更新
		if sys.httptime.After(sys.pupdatetime) {
			log.Println("time error")
			return
		}

		cel, ok := sys.CellMap.Load(instId)
		if !ok {
			log.Println("数据错误")
			return
		}
		cell := (cel.(*Cell))
		//查询仓位
		cpos := sys.getPositionBysymbol(instId)
		if sys.finesse(cell, cpos) {
			sys.httptime = time.Now()
			sys.resettimer()
		}

	})

	sys.net.Rigist(emptyactor, 1)

}

//策略执行
func (sys *Sys) finesse(cell *Cell, pos *entity.Position) bool {
	log.Println("finnese")
	servobj := sys.fClient.NewCreateOrderService().Symbol(cell.brd.InstId).Type(futures.OrderTypeMarket)
	tdc, ma7, lastc := cell.brd.GetTdc()
	if lastc == 0 {
		log.Println("close 0")
		return false
	}
	bios := (lastc - ma7) * 100 / lastc
	log.Println("bios:", bios)
	if pos != nil { //已有仓位

		servobj.PositionSide(futures.PositionSideType(pos.Side))

		//正常平仓

		isclose := false

		//平仓条件
		if pos.Size < 0 && lastc < ma7 {
			isclose = true
		}
		if pos.Size > 0 && lastc > ma7 {
			isclose = true
		}

		//止损平仓
		if isclose || pos.UnPnl+cell.Cnf.Maxloss < 0 {
			if pos.Size < 0 {
				servobj.Side(futures.SideTypeBuy)
			} else {
				servobj.Side(futures.SideTypeSell)
			}

			servobj.Quantity(fmt.Sprintf("%f", math.Abs(pos.Size)))
			_, err := servobj.Do(context.Background())
			if err != nil {
				log.Println(err.Error())
				return false
			} else {
				return true
			}
		}

		//加仓
		if pos.Persont < cell.Cnf.Addposition && math.Abs(pos.Size)*2 < cell.Cnf.Maxcount {
			if pos.Size < 0 {
				servobj.Side(futures.SideTypeSell)
			} else {
				servobj.Side(futures.SideTypeBuy)
			}
			servobj.Quantity(fmt.Sprintf("%f", math.Abs(pos.Size)))
			_, err := servobj.Do(context.Background())
			if err != nil {
				log.Println(err.Error())
				return false
			} else {
				return true
			}

		}

	} else {
		if tdc >= 13 && bios > float64(0.5) || tdc <= -13 && bios < float64(-0.5) {
			if tdc > 0 {
				servobj.PositionSide(futures.PositionSideTypeShort)
				servobj.Side(futures.SideTypeSell)
			} else {
				servobj.PositionSide(futures.PositionSideTypeLong)
				servobj.Side(futures.SideTypeBuy)
			}
			quantity := fmt.Sprintf("%f", cell.Cnf.Firstcount)
			log.Println("quantity:", quantity)
			servobj.Quantity(quantity)
			_, err := servobj.Do(context.Background())
			if err != nil {
				log.Println(err.Error())
				return false
			} else {
				return true
			}
		}

	}

	return false
}
