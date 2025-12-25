// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"

	"hello-gozero/internal/config"
	"hello-gozero/internal/routes"
	"hello-gozero/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/hellogozero.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		fmt.Printf("failed to create service context: %v\n", err)
		return
	}
	defer ctx.Close()
	routes.RegisterHandlers(server, ctx)

	fmt.Printf("🚀 Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
