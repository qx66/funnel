package sink

import (
	"github.com/qx66/funnel/internal/biz"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"time"
)

import (
	"context"
)

type KafkaDataSource struct {
	kafka *Kafka
}

func NewKafkaSink(kafka *Kafka) biz.SinkRepo {
	return &KafkaDataSource{
		kafka: kafka,
	}
}

func (kafkaDataSource *KafkaDataSource) Product(ctx context.Context) {
	var msgs []kafka.Message
	ticker := time.Tick(1 * time.Second)
	
	for {
		select {
		case msg := <-*kafkaDataSource.kafka.message:
			msgs = append(msgs, msg)
			if len(msgs) >= 30 {
				newMsgs := make([]kafka.Message, len(msgs))
				copy(newMsgs, msgs)
				msgs = msgs[:0]
				go kafkaDataSource.Record(ctx, newMsgs...)
			}
		
		case <-ticker:
			if len(msgs) > 0 {
				newMsgs := make([]kafka.Message, len(msgs))
				copy(newMsgs, msgs)
				msgs = msgs[:0]
				go kafkaDataSource.Record(ctx, newMsgs...)
			}
		
		case <-ctx.Done():
			if len(msgs) > 0 {
				kafkaDataSource.Record(ctx, msgs...)
			}
			return
		}
	}
}

func (kafkaDataSource *KafkaDataSource) Record(ctx context.Context, msgs ...kafka.Message) error {
	err := kafkaDataSource.kafka.writer.WriteMessages(ctx, msgs...)
	if err != nil {
		kafkaDataSource.kafka.logger.Error(
			"插入数据失败",
			zap.Error(err),
			zap.Any("msgs", msgs),
		)
	}
	return err
}
