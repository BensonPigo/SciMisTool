package mq

import (
	config "TpeBiConsumer/config"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type MQClient struct {
	cfg      config.MQConfig
	conn     *amqp.Connection
	ch       *amqp.Channel
	queue    *amqp.Queue
	confirms <-chan amqp.Confirmation
	mu       sync.Mutex
}

// Consumer 依 RoutingKey 分流，並支援優雅關閉
type Consumer struct {
	client    *MQClient
	ch        *amqp.Channel
	queueName string
	logger    *zap.SugaredLogger
	wg        sync.WaitGroup
	mu        sync.Mutex
}

// ensureChannel creates or reuses the consumer's channel safely
func (c *Consumer) ensureChannel() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ch != nil && !c.ch.IsClosed() {
		return nil
	}

	ch, err := c.client.conn.Channel()
	if err != nil {
		return err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		return err
	}
	c.ch = ch
	c.queueName = c.client.queue.Name
	return nil
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
		cfg.DeadLetterExchange, // DLX 的 Exchange 名稱
		"direct",               // 使用 direct 類型
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("宣告 Dead Letter Exchange 失敗: %w", err)
	}
	// =================================================
	// 2. 宣告 Dead Letter Queue 並綁定到上面的 DLX（dead_ddl_queue）
	// =================================================
	deadQ, err := ch.QueueDeclare(
		cfg.DeadLetterQueue, // DLQ 名稱
		true,                // durable
		false,               // auto-delete
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("宣告 Dead Letter Queue 失敗: %w", err)
	}
	// 將 dead_ddl_queue 綁定到 ddl_dlx_exchange，routing key 為 "dead_ddldml"
	err = ch.QueueBind(
		deadQ.Name,
		cfg.DeadLetterRoutingKey, // Dead Letter RoutingKey
		cfg.DeadLetterExchange,   // Dead Letter Exchange 名稱
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
		cfg.PrimaryExchange, // Exchange name
		"direct",            // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("宣告 Primary Exchange 失敗: %w", err)
	}

	// =====================================================
	// 4. 宣告 Primary Queue（ddl_dml_main_queue），並加上 Dead Letter 設定
	// =====================================================
	//    當 consumer Nack(false,false) 時，訊息會自動透過以下參數送到 ddldml_dlx_exchange
	args := amqp.Table{
		"x-dead-letter-exchange":    cfg.DeadLetterExchange,
		"x-dead-letter-routing-key": cfg.DeadLetterRoutingKey,
	}
	q, err := ch.QueueDeclare(
		cfg.PrimaryQueue, // queue 名稱
		true,             // durable
		false,            // auto-delete
		false,            // exclusive
		false,            // no-wait
		args,             // 這裡帶入 DLX 參數
	)
	if err != nil {
		return nil, fmt.Errorf("宣告 Primary Queue 失敗: %w", err)
	}

	// Bind 多個 RoutingKey 到這個 Queue
	for _, routingKey := range AllRoutingKeys {
		err := ch.QueueBind(
			q.Name,
			string(routingKey), // 注意要轉成 string
			cfg.PrimaryExchange,
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

// Reconnect 重新建立與 RabbitMQ 的連線與相關資源
func (c *MQClient) Reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && !c.conn.IsClosed() {
		// another goroutine already reconnected
		return nil
	}

	newClient, err := NewMQClient(c.cfg)
	if err != nil {
		return err
	}

	if c.conn != nil {
		c.Close()
	}
	c.conn = newClient.conn
	c.ch = newClient.ch
	c.queue = newClient.queue
	c.confirms = newClient.confirms
	return nil
}

// NewConsumer 建立一個 consumer
func NewConsumer(mqClient *MQClient, logger *zap.SugaredLogger) *Consumer {
	c := &Consumer{
		client: mqClient,
		logger: logger,
	}
	if err := c.ensureChannel(); err != nil {
		return nil
	}
	return c
}

// reconnect 重新連線並建立新的 channel
func (c *Consumer) reconnect(ctx context.Context) error {
	for {
		if err := c.client.Reconnect(); err != nil {
			c.logger.Errorw("MQ reconnect failed", "err", err)
		} else if err := c.ensureChannel(); err == nil {
			return nil
		} else {
			c.logger.Errorw("Channel open failed", "err", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}

// Start 啟動消費，並在收到 ctx.Done() 時優雅結束
func (c *Consumer) Start(ctx context.Context, handler HandlerFunc) error {
	if err := c.ensureChannel(); err != nil {
		return err
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			msgs, err := c.ch.Consume(
				c.queueName,
				"",
				false,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				c.logger.Errorw("Consume failed", "err", err)
				if err := c.reconnect(ctx); err != nil {
					c.logger.Errorw("Reconnect failed", "err", err)
					return
				}
				continue
			}

			for {
				select {
				case <-ctx.Done():
					c.logger.Info("Consumer received shutdown signal")
					return

				case d, ok := <-msgs:
					if !ok {
						if err := c.reconnect(ctx); err != nil {
							c.logger.Errorw("Reconnect failed", "err", err)
							return
						}
						break
					}
					if err := handler(ctx, d.RoutingKey, d.Body); err != nil {
						c.logger.Errorw("Handler error, message will be sent to Dead letter queue",
							"routingKey", d.RoutingKey,
							"err", err,
						)
						d.Nack(false, false)
					} else {
						d.Ack(false)
					}
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
