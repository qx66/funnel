package main

import (
	"bytes"
	"context"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/qx66/funnel/internal/biz"
	"github.com/qx66/funnel/internal/conf"
	"github.com/qx66/funnel/pkg/middleware"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type app struct {
	transfer *biz.Transfer
}

func newApp(transfer *biz.Transfer) *app {
	return &app{
		transfer: transfer,
	}
}

var configPath string

func init() {
	flag.StringVar(&configPath, "configPath", "", "-configPath")
}

func main() {
	flag.Parse()
	
	logger, _ := zap.NewProduction(zap.Fields(zap.String("service", "funnel")))
	defer logger.Sync()
	
	if configPath == "" {
		logger.Error("configPath 参数为空")
		return
	}
	
	//
	f, err := os.Open(configPath)
	defer f.Close()
	if err != nil {
		logger.Error(
			"加载配置文件失败",
			zap.String("configPath", configPath),
			zap.Error(err),
		)
		return
	}
	
	//
	var buf bytes.Buffer
	_, err = io.Copy(&buf, f)
	if err != nil {
		logger.Error(
			"加载配置文件copy内容失败",
			zap.Error(err),
		)
		return
	}
	
	//
	var bootstrap conf.Bootstrap
	err = yaml.Unmarshal(buf.Bytes(), &bootstrap)
	if err != nil {
		logger.Error(
			"序列化配置失败",
			zap.Error(err),
		)
		return
	}
	
	message := make(chan kafka.Message)
	app, err := initApp(&bootstrap, &message, logger)
	
	//
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go app.transfer.Product(ctx)
	
	//
	r := gin.New()
	r.Use(middleware.Recording(logger))
	r.GET("/*filepath", app.transfer.Funnel)
	r.POST("/*filepath", app.transfer.Funnel)
	
	err = r.Run(":16666")
	if err != nil {
		logger.Error(
			"启动 funnel 服务失败",
			zap.Error(err),
		)
	}
}
