package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/chaosknight/biantrade/entity"
	"github.com/chaosknight/go-binance/v2/futures"
	"github.com/chaosknight/skynet/skynet"
	"github.com/chaosknight/skynet/types"
)

const timestep = 10 * time.Second

type Sys struct {
	cnf         *entity.Sysconfig
	listenKey   string
	fClient     *futures.Client
	positions   []*entity.Position
	pupdatetime time.Time
	httptime    time.Time
	accwsch     chan struct{}
	klinech     chan struct{}
	net         *skynet.SkyNet
	pmux        sync.Mutex
	CellMap     sync.Map
	timer       *time.Timer //超时还未收到仓位推送，主动拉取
}

func NewSys(cnf *entity.Sysconfig) *Sys {
	apikey, scrkey := cnf.GetUserKey()

	sys := &Sys{
		cnf:       cnf,
		net:       &skynet.SkyNet{},
		fClient:   futures.NewClient(apikey, scrkey),
		positions: []*entity.Position{},
	}
	sys.httptime = time.Now()
	sys.pupdatetime = time.Now()

	sys.net.Init(types.SkyNetInitOptions{})

	return sys
}

func (sys *Sys) init() {
	for _, v := range sys.cnf.Insts {
		sys.CellMap.Store(v.InstID, NewCell(v, "5m"))
	}
	sys.actorinit()
}

func (sys *Sys) Start() {
	if sys.getlistenkey() != nil {
		return
	}

	if sys.getAccount() != nil {
		return
	}

	sys.init()
	sys.setAccServe()
	sys.setKlineServe()

	select {
	case <-sys.accwsch:
		sys.setAccServe()
	case <-sys.klinech:
		sys.setKlineServe()
	}

}

func (sys *Sys) setKlineServe() {
	errHandler := func(err error) {
		fmt.Println(err)
	}

	mpkeys := map[string]string{}

	for _, v := range sys.cnf.Insts {
		mpkeys[v.InstID] = "5m"
	}

	doneC, _, err := futures.WsCombinedKlineServe(mpkeys, sys.wsKlineHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	sys.klinech = doneC
}
func (sys *Sys) wsKlineHandler(event *futures.WsKlineEvent) {
	instId := event.Symbol

	log.Println(event.Symbol, event.Kline.Close)
	//更新仓位信息，计算盈亏
	sys.pmux.Lock()
	defer sys.pmux.Unlock()
	cpos := sys.getPositionBysymbol(instId)
	if cpos != nil {
		v, err := strconv.ParseFloat(event.Kline.Close, 64)
		if err != nil {
			return
		}
		//计算盈亏
		cpos.SetNewPrice(v)
		log.Println(cpos)
	}
	//更新k线信息
	v, ok := sys.CellMap.Load(instId)
	if !ok {
		//
	} else {
		cell := v.(*Cell)
		if cell.brd.IsEmpty() { //加载k线
			sys.net.SendMsg(LoadBreed, instId)
		} else {
			cell.brd.SetNewK(&event.Kline)
		}
		//发送消息，执行策略
		sys.net.SendMsg(OrderActor, event.Symbol)
	}

}

func (sys *Sys) setAccServe() {
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, _, err := futures.WsUserDataServe(sys.listenKey, sys.wsaccHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	sys.accwsch = doneC
}

func (sys *Sys) wsaccHandler(event *futures.WsUserDataEvent) {
	switch event.Event {
	//key 过期
	case futures.UserDataEventTypeListenKeyExpired:
		sys.keeplistenkey()
	//追加保证金
	case futures.UserDataEventTypeMarginCall:
	case futures.UserDataEventTypeAccountUpdate:
		sys.pmux.Lock()
		sys.positions = wstoposition(event.AccountUpdate.Positions)
		sys.pupdatetime = time.Now()
		if sys.timer != nil {
			sys.timer.Stop()
		}
		sys.pmux.Unlock()
		log.Println(sys.positions)
	case futures.UserDataEventTypeOrderTradeUpdate:
	case futures.UserDataEventTypeAccountConfigUpdate:
	}
}

func (sys *Sys) getAccount() error {
	acc, err := sys.fClient.NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Println("获取账户信息失败")
		return err

	}
	sys.pmux.Lock()
	sys.positions = acctoposition(acc.Positions)
	sys.pupdatetime = time.Now()
	sys.pmux.Unlock()
	log.Println(sys.positions)
	return nil
}

func (sys *Sys) getlistenkey() error {
	lkey, err := sys.fClient.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		fmt.Println("获取listenKey失败")
		return err
	}
	sys.listenKey = lkey
	return nil
}

func (sys *Sys) keeplistenkey() error {
	return sys.fClient.NewKeepaliveUserStreamService().ListenKey(sys.listenKey).Do(context.Background())
}

//查询仓位
func (sys *Sys) getPositionBysymbol(symbol string) *entity.Position {
	for _, v := range sys.positions {
		if v.Symbol == symbol {
			return v
		}
	}
	return nil
}

func (sys *Sys) resettimer() {
	if sys.timer == nil {
		sys.timer = time.NewTimer(timestep)
		go func() {
			for true {
				<-sys.timer.C
				if sys.httptime.After(sys.pupdatetime) {
					if sys.getAccount() != nil {
						sys.timer.Reset(timestep)
					}
				}

			}

		}()
	} else {
		sys.timer.Reset(timestep)
	}
}

//accountposition to position
func acctoposition(accpos []*futures.AccountPosition) (pos []*entity.Position) {
	if len(accpos) > 0 {
		for _, p := range accpos {
			if p.PositionAmt == 0 {
				continue
			}
			pos = append(pos, &entity.Position{
				Symbol: p.Symbol,
				Size:   float64(p.PositionAmt),
				Side:   string(p.PositionSide),
				Price:  float64(p.EntryPrice),
			})
		}
	}
	return
}

//wsposition to postion
func wstoposition(wspos []futures.WsPosition) (pos []*entity.Position) {
	if len(wspos) > 0 {
		for _, p := range wspos {
			if p.Amount == 0 {
				continue
			}
			pos = append(pos, &entity.Position{
				Symbol: p.Symbol,
				Size:   float64(p.Amount),
				Side:   string(p.Side),
				Price:  float64(p.EntryPrice),
			})
		}
	}
	return
}
