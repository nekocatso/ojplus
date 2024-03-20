package routers

import (
	"Alarm/internal/web/controllers"
)

func (router *Router) AuthInit(ctrl *controllers.Auth) {
	group := router.engine.Group("")
	{
		group.POST("/login", ctrl.Login)
		group.POST("/register", ctrl.Register)
	}
}
