package main

import (
	"gitlab.xfq.com/tech-lab/dionysus"
	"gitlab.xfq.com/tech-lab/dionysus/cmd/gincmd"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/conf"
	dredis "gitlab.xfq.com/tech-lab/dionysus/pkg/redis"
	"go.uber.org/dig"
	"log"
	itemsHandler "tbTool/api/handler/items"
	"tbTool/api/routers"
	"tbTool/api/service/items"
)

func initContainer() *dig.Container {
	c := dig.New()

	itemSrvErr := c.Provide(items.NewItemServiceImpl)
	itemHandErr := c.Provide(itemsHandler.NewItemsOnSaleGetHandler)
	log.Fatalf("initContainer start items result:%v,%v", itemSrvErr, itemHandErr)

	return c
}

func main() {
	g := gincmd.New()

	err := g.RegPreRunFunc("watch.redis", 1, func() error {
		return conf.RegisterEtcdWatch(&dredis.RedisEvent{Prefix: "watch.redis"})
	})
	if err != nil {
		log.Println("Reg pre run func err:", err)
	}

	err = g.RegPreRunFunc("business", 2, func() error {
		return conf.StartWatchConfig("business")
	})
	if err != nil {
		log.Println("Reg pre run func err:", err)
	}

	err = g.RegPreRunFunc("watch.mysql", 3, func() error {
		return conf.StartWatchConfig("watch.mysql")
	})
	if err != nil {
		log.Println("Reg pre run func err:", err)
	}
	
	_ = g.RegPreRunFunc("initContainer", 5, func() error {
		//依赖注入
		log.Println("initContainer start")
		c := initContainer()

		//路由注入
		log.Println("RegisterRouter start step")
		routers.RegisterRouter(c, g.Engine)
		return nil
	})

	dionysus.Start("gapi", g)
}
