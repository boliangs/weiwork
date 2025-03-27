package main

import (
	"sendswork/app/gateway/internal"
	"sendswork/app/gateway/router"
	"sendswork/app/gateway/rpc"
	"sendswork/config"
)

func main() {
	config.InitConfig()
	rpc.Init()
	r := router.Router()
	//go internal.SendOnlineCount()
	go internal.YearBillDataInitSync()
	r.Run("0.0.0.0:8889")
}
