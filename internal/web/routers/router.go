package routers

import "github.com/gin-gonic/gin"

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
