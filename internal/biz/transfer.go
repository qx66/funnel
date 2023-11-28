package biz

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/wire"
	"github.com/qx66/funnel/internal/conf"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"strings"
	"time"
)

type Transfer struct {
	topic    string
	sinkRepo SinkRepo
	message  *chan kafka.Message
	logger   *zap.Logger
}

type SinkRepo interface {
	Product(ctx context.Context)
	// Record(ctx context.Context, msgs ...kafka.Message) error
}

func NewTransfer(bootstrap *conf.Bootstrap, sinkRepo SinkRepo, message *chan kafka.Message, logger *zap.Logger) *Transfer {
	if bootstrap.Sinks.Topic == "" {
		panic("topic 不能为空")
	}
	
	return &Transfer{
		topic:    bootstrap.Sinks.Topic,
		sinkRepo: sinkRepo,
		message:  message,
		logger:   logger,
	}
}

var ProviderSet = wire.NewSet(NewTransfer)

func (transfer *Transfer) Product(ctx context.Context) {
	transfer.sinkRepo.Product(ctx)
}

func (transfer *Transfer) Funnel(c *gin.Context) {
	doc := make(map[string]any)
	
	// basic
	doc["timestamp"] = time.Now().Unix()
	doc["method"] = c.Request.Method
	doc["host"] = c.Request.Host
	doc["requestUri"] = c.Request.RequestURI
	doc["remoteAddr"] = c.Request.RemoteAddr
	doc["urlPath"] = c.Request.URL.Path
	// header
	for k, v := range c.Request.Header {
		doc[strings.Join([]string{"header", k}, "-")] = strings.Join(v, " ")
	}
	
	// cookie
	cookies := c.Request.Cookies()
	for _, cookie := range cookies {
		doc[strings.Join([]string{"cookie", cookie.Name}, "-")] = cookie.Value
	}
	
	// get query
	reqParams := c.Request.URL.Query()
	for k, v := range reqParams {
		doc[strings.Join([]string{"query", k}, "-")] = strings.Join(v, " ")
	}
	
	// post - 只对 Json 和 Form 进行处理
	
	if c.Request.Method == "POST" {
		contentType := c.ContentType()
		
		switch contentType {
		case binding.MIMEJSON:
			rawDataMap := make(map[string]interface{})
			rawDataByte, err := c.GetRawData()
			
			if len(rawDataByte) != 0 {
				err = json.Unmarshal(rawDataByte, &rawDataMap)
				if err != nil {
					transfer.logger.Error(
						"解析json数据失败",
						zap.Error(err),
					)
					break
				}
				
				for k, v := range rawDataMap {
					doc[strings.Join([]string{"json", k}, "-")] = v
				}
			}
			
			// Form 数据会自动叠加 Get Query 数据
		case binding.MIMEPOSTForm:
			err := c.Request.ParseForm()
			if err != nil {
				transfer.logger.Error(
					"解析Form数据失败",
					zap.Error(err),
				)
				break
			}
			form := c.Request.Form
			for k, v := range form {
				doc[strings.Join([]string{"form", k}, "-")] = strings.Join(v, " ")
			}
			
			/*
				case binding.MIMEXML, binding.MIMEXML2:
					c.JSON(400, gin.H{"code": 10400})
					return
				case binding.MIMEPROTOBUF:
					c.JSON(400, gin.H{"code": 10400})
					return
				case binding.MIMEMSGPACK, binding.MIMEMSGPACK2:
					c.JSON(400, gin.H{"code": 10400})
					return
				case binding.MIMEYAML:
					c.JSON(400, gin.H{"code": 10400})
					return
				case binding.MIMETOML:
					c.JSON(400, gin.H{"code": 10400})
					return
				case binding.MIMEMultipartPOSTForm:
					c.JSON(400, gin.H{"code": 10400})
					return
			*/
		default:
			transfer.logger.Error(
				"解析Post数据，暂时不支持该内容类型",
				zap.String("contentType", contentType),
			)
			
			c.JSON(400, gin.H{"code": 10400})
			return
		}
	}
	
	// doc json
	docByte, err := json.Marshal(doc)
	if err != nil {
		transfer.logger.Error(
			"json序列化结果失败",
			zap.Error(err),
		)
		c.JSON(500, gin.H{"code": 10500})
		return
	}
	
	//
	*transfer.message <- kafka.Message{
		Topic: transfer.topic,
		Value: docByte,
	}
	
	c.JSON(200, gin.H{"code": 0})
	return
}
