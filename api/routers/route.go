package routers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
	"log"
	"tbTool/api/handler/items"
	. "tbTool/api/middleware"
)

func RegisterRouter(c *dig.Container, e *gin.Engine) {

	api := e.Group("/tbApi")
	api.Use(Sign())

	if err := c.Invoke(func(h *items.ItemOnSaleGetHandler) {

		api.POST("items/ItemsOnSaleGet", func(ctx *gin.Context) { ctx.Render(200, h.TaoBaoItemsOnSaleGet(ctx)) })

	}); err != nil {
		log.Fatalf("%s", err)
	}

}
