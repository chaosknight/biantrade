package main

import (
	"fmt"

	"github.com/chaosknight/biantrade/entity"
)

func main() {

	syscnf := entity.LoadCnf("./cfg.json")
	if syscnf == nil {
		fmt.Println("配置文件错误")
		return
	}

	sys := NewSys(syscnf)
	sys.Start()

}
