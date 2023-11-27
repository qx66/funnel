package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction(zap.Fields(zap.String("service", "funnel")))
	defer logger.Sync()
	
	r := gin.New()
	
	r.GET("/*filepath")
	
	err := r.Run(":16666")
	if err != nil {
	
	}
}
