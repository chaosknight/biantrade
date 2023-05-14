package entity

import (
	"bytes"
	"encoding/json"
	"os"
)

type FinesseConfig struct {
	InstID string `json:"instId"`
	//底仓
	Firstcount float64 `json:"firstcount"`
	//最大仓位
	Maxcount float64 `json:"maxcount"`
	//加仓条件
	Addposition float64 `json:"addposition"`
	//最大亏损 触发平仓
	Maxloss float64 `json:"maxloss"`
}

type Sysconfig struct {
	User [3]string `json:"user"`

	Insts []*FinesseConfig `json:"insts"`
}

func (cnf *Sysconfig) GetUserKey() (apikey, skey string) {
	return cnf.User[0], cnf.User[2]
}

func (cnf *Sysconfig) GetConfigByid(instid string) *FinesseConfig {
	for _, v := range cnf.Insts {
		if v.InstID == instid {
			return v
		}
	}
	return nil
}

func LoadCnf(fname string) *Sysconfig {
	f, err := os.Open(fname)
	if err != nil {
		return nil
	}
	defer f.Close()
	b := new(bytes.Buffer)
	_, err = b.ReadFrom(f)
	if err != nil {
		return nil
	}
	sf := Sysconfig{}
	err = json.Unmarshal(b.Bytes(), &sf)

	if err != nil {
		return nil
	}

	return &sf
}
