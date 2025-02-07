package main

import (
	"ChatRoomAPI/src"
	"ChatRoomAPI/src/controller"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	root := gin.New()
	root.SetTrustedProxies([]string{"192.168.1.1", "127.0.0.1"})
	apiv1 := root.Group("/api/v1")
	controller.MiddlewareInit(apiv1)
	root.Run(fmt.Sprintf(":%d", src.GlobalConfig.YamlConfig.Server.Port))
}
