package main

import (
	"idea-go/bootstrap"
	router "idea-go/routes"
)

const ServiceHostPort = ":50069"

func main() {
	// 初始化Web
	bootstrap.Init()
	r := bootstrap.InitWeb()
	router.RegisterRouter(r)
	bootstrap.RunWeb(r, ServiceHostPort)
}
