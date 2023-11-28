//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/qx66/funnel/internal/biz"
	"github.com/qx66/funnel/internal/conf"
	"github.com/qx66/funnel/internal/sink"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func initApp(*conf.Bootstrap, *chan kafka.Message, *zap.Logger) (*app, error) {
	panic(wire.Build(
		sink.ProviderSet,
		biz.ProviderSet,
		newApp))
}
