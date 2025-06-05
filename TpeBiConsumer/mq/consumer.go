package mq

import (
	config "TpeBiConsumer/config"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type MQClient struct {
	cfg      config.MQConfig
	conn     *amqp.Connection
	ch       *amqp.Channel
	queue    *amqp.Queue
	confirms <-chan amqp.Confirmation
}

// Consumer 依 RoutingKey 分流，並支援優雅關閉
type Consumer struct {
	ch        *amqp.Channel
	queueName string
	logger    *zap.SugaredLogger
	wg        sync.WaitGroup
}

// HandlerFunc 處理訊息的 callback
// routingKey: 來源
// body: 訊息內容
type HandlerFunc func(ctx context.Context, routingKey string, body []byte) error

// 建立 RabbitMQ TLS、連線、Channel、宣告 Exchange...
func NewMQClient(cfg config.MQConfig) (*MQClient, error) {

	// 載入客戶端憑證與私鑰
	clientCert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("載入客戶端憑證失敗: %w", err)
	}

	// 載入信任的 CA 憑證
	caCert, err := os.ReadFile(cfg.CACertFile)
	if err != nil {
		return nil, fmt.Errorf("讀取 CA 憑證失敗: %w", err)
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("加入 CA 憑證失敗")
	}

	// 建立 TLS 配置
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false, // 正式環境建議驗證伺服器憑證
	}

	// 使用 TLS 建立連線，連線字串改為 amqps 並連到 5671 端口
	conn, err := amqp.DialTLS(cfg.AMQPURL, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("連線 RabbitMQ 失敗: %w", err)
	}

	// Channel 宣告
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("開啟 Channel 失敗: %w", err)
	}

	// Channel 開啟 Confirm 模式
	if err := ch.Confirm(false); err != nil {
		return nil, fmt.Errorf("開啟 Confirm 模式失敗：%w", err)
	}

	// 建立一個有緩衝的確認 channel
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	/*------------------Dead Letter Queue------------------*/

	// ============================
	// 1. 宣告 Dead Letter Exchange
	// ============================
	err = ch.ExchangeDeclare(
		"ddl_dlx_exchange", // DLX 的 Exchange 名稱
		"direct",           // 使用 direct 類型
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("宣告 Dead Letter Exchange 失敗: %w", err)
	}
	// =================================================
	// 2. 宣告 Dead Letter Queue 並綁定到上面的 DLX（dead_ddl_queue）
	// =================================================
	deadQ, err := ch.QueueDeclare(
		"ddl_dead_queue", // DLQ 名稱
		true,             // durable
		false,            // auto-delete
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("宣告 Dead Letter Queue 失敗: %w", err)
	}
	// 將 dead_ddl_queue 綁定到 ddl_dlx_exchange，routing key 為 "dead_ddl"
	err = ch.QueueBind(
		deadQ.Name,
		"dead_ddl",         // Dead Letter RoutingKey
		"ddl_dlx_exchange", // Dead Letter Exchange 名稱
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("綁定 Dead Letter Queue 失敗: %w", err)
	}
	// ============================
	// 3. 宣告 Primary Exchange（global_direct）
	// ============================
	err = ch.ExchangeDeclare(
		"global_direct", // Exchange name
		"direct",        // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("宣告 Primary Exchange 失敗: %w", err)
	}

	// =====================================================
	// 4. 宣告 Primary Queue（ddl_queue），並加上 Dead Letter 設定
	// =====================================================
	//    當 consumer Nack(false,false) 時，訊息會自動透過以下參數送到 ddl_dlx_exchange
	args := amqp.Table{
		"x-dead-letter-exchange":    "ddl_dlx_exchange", // 指定死信 Exchange
		"x-dead-letter-routing-key": "dead_ddl",         // 指定 routing key
	}
	q, err := ch.QueueDeclare(
		"ddl_queue", // queue 名稱
		true,        // durable
		false,       // auto-delete
		false,       // exclusive
		false,       // no-wait
		args,        // 這裡帶入 DLX 參數
	)
	if err != nil {
		return nil, fmt.Errorf("宣告 Primary Queue 失敗: %w", err)
	}

	/*------------------Dead Letter Queue------------------*/

	// Exchange 宣告
	// err = ch.ExchangeDeclare(
	// 	"global_direct", // Exchange name
	// 	"direct",        // type
	// 	true,            // durable
	// 	false,           // auto-deleted
	// 	false,           // internal
	// 	false,           // no-wait
	// 	nil,             // arguments
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf("宣告 Exchange 失敗: %w", err)
	// }

	// Queue 宣告
	// q, err := ch.QueueDeclare(
	// 	"ddl_queue", // queue 名稱
	// 	true,        // durable
	// 	false,       // auto-delete
	// 	false,       // exclusive
	// 	false,       // no-wait
	// 	nil,         // arguments
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf("開啟 Channel 失敗: %w", err)
	// }

	// Bind 多個 RoutingKey 到這個 Queue
	for _, routingKey := range AllRoutingKeys {
		err := ch.QueueBind(
			q.Name,
			string(routingKey), // 注意要轉成 string
			"global_direct",
			false, nil,
		)
		if err != nil {
			return nil, fmt.Errorf("queue RoutingKey Bind(%s) 失敗：%v", string(routingKey), err)
		}
	}

	return &MQClient{conn: conn, ch: ch, cfg: cfg, queue: &q, confirms: confirms}, nil
}

func (c *MQClient) Close() {
	c.ch.Close()
	c.conn.Close()
}

// NewConsumer 建立一個 consumer
func NewConsumer(mqClient *MQClient, logger *zap.SugaredLogger) *Consumer {
	ch, err := mqClient.conn.Channel() // ← 每個 consumer 自己的 channel
	if err != nil {
		return nil
	}
	if err := ch.Qos(1, 0, false); err != nil {
		return nil
	}

	return &Consumer{
		ch:        ch,
		queueName: mqClient.queue.Name,
		logger:    logger,
	}
}

// Start 啟動消費，並在收到 ctx.Done() 時優雅結束
func (c *Consumer) Start(ctx context.Context, handler HandlerFunc) error {
	msgs, err := c.ch.Consume(
		c.queueName,
		"",    // consumerTag：空字串讓 RabbitMQ 自動產生唯一 tag
		false, // autoAck = false，由我們手動 Ack / Nack
		false, // exclusive = false，允許多個 consumer 共同消費同一個 queue
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Consumer received shutdown signal")
				return

			case d, ok := <-msgs:
				if !ok {
					c.logger.Warn("Delivery channel closed")
					return
				}
				// 呼叫外部 Handler
				if err := handler(ctx, d.RoutingKey, d.Body); err != nil {

					// 處理失敗時，Nack 並設定 requeue = false
					// 使此筆訊息被送到我們在 QueueDeclare 時指定的 Dead Letter Exchange
					c.logger.Errorw("Handler error, message will be sent to Dead letter queue",
						"routingKey", d.RoutingKey,
						"err", err,
					)
					d.Nack(false, false) // requeue = false → 走死信機制
					/*c.logger.Errorw("Message處理失敗", "routingKey", d.RoutingKey, "err", err)
					// 處理失敗，只針對這筆消息回應 Nack，該消息則重新回到 queue
					d.Nack(false, true)*/
				} else {
					// 處理成功，只針對這筆消息回應 ack
					d.Ack(false)
				}
			}
		}
	}()
	return nil
}

// Wait 等待所有 goroutine 完成
func (c *Consumer) Wait() {
	c.wg.Wait()
}
