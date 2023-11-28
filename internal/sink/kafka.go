package sink

import (
	"github.com/google/wire"
	"github.com/qx66/funnel/internal/conf"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"time"
)

type Kafka struct {
	writer  *kafka.Writer
	message *chan kafka.Message
	logger  *zap.Logger
}

var ProviderSet = wire.NewSet(NewMQ, NewKafkaSink)

func NewMQ(bootstrap *conf.Bootstrap, message *chan kafka.Message, logger *zap.Logger) *Kafka {
	if len(bootstrap.Sinks.Kafka.BootstrapServer) == 0 {
		panic("kafka 地址不能为空")
	}
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(bootstrap.Sinks.Kafka.BootstrapServer...),
		Balancer:               &kafka.LeastBytes{},
		BatchTimeout:           time.Duration(bootstrap.Sinks.Kafka.BatchTimeout) * time.Millisecond,
		ReadTimeout:            time.Duration(bootstrap.Sinks.Kafka.ReadTimeout) * time.Millisecond,
		WriteTimeout:           time.Duration(bootstrap.Sinks.Kafka.WriteTimeout) * time.Millisecond,
		AllowAutoTopicCreation: true,
	}
	
	return &Kafka{
		writer:  writer,
		message: message,
		logger:  logger,
	}
}
