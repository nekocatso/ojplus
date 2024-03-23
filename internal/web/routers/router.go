package routers

import (
	"Alarm/internal/web/controllers"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter() *Router {
	return &Router{
		engine: gin.Default(),
	}
}

func (r *Router) Run(addr ...string) {
	r.engine.Run(addr...)
}

func (r *Router) AccountCtrl(ctrl *controllers.Account) {
	group := r.engine.Group("")
	{
		group.POST("/register", ctrl.CreateUser)
		group.GET("/users", ctrl.AllUser)
		group.GET("/users/:id", ctrl.GetUserByID)
		group.PATCH("/users/:id", ctrl.UpdateUserByID)
	}
}
func (r *Router) AuthCtrl(ctrl *controllers.Auth) {
	group := r.engine.Group("")
	{
		group.POST("/login", ctrl.Login)
		group.POST("/logout", ctrl.Logout)
	}
}
