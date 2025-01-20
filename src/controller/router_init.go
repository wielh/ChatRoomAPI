package controller

import (
	"github.com/gin-gonic/gin"
)

func MiddlewareInit(g *gin.RouterGroup) {
	commonMiddleware(g)
	userGroupRouter(g)
	roomGroupRouter(g)
	messageGroupRouter(g)
}
