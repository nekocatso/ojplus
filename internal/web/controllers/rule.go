package controllers

import (
	"Alarm/internal/web/services"

	"github.com/gin-gonic/gin"
)

type Rule struct {
	svc *services.Rule
	cfg map[string]interface{}
}

func NewRule(cfg map[string]interface{}) *Rule {
	svc := services.NewRule(cfg)
	return &Rule{svc: svc, cfg: cfg}
}

func (ctrl *Rule) CreateRule(ctx *gin.Context) {

}
