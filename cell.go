package main

import (
	"github.com/chaosknight/biantrade/entity"
)

type Cell struct {
	Cnf *entity.FinesseConfig
	brd *Breed
}

func NewCell(cnf *entity.FinesseConfig, interval string) *Cell {
	return &Cell{
		Cnf: cnf,

		brd: NewBreed(cnf.InstID, interval),
	}
}
