// 發送 RabbitMQ 訊息
package mq

import (
	config "FtyBiProducer/config"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MQClient struct {
	cfg      config.MQConfig
	conn     *amqp.Connection
	ch       *amqp.Channel
	queue    *amqp.Queue
	confirms <-chan amqp.Confirmation
	mu       sync.Mutex
}

// ensureChannel creates or reuses the producer's channel safely
func (c *MQClient) ensureChannel() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ch != nil && !c.ch.IsClosed() {
		return nil
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		return err
	}
	c.ch = ch
	c.confirms = ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	return nil
}

// Reconnect re-establishes the connection and channel if needed
func (c *MQClient) Reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && !c.conn.IsClosed() {
		// already connected by another goroutine
		return nil
	}

	newClient, err := NewMQClient(c.cfg)
	if err != nil {
		return err
	}

	if c.conn != nil {
		c.ch.Close()
		c.conn.Close()
	}
	c.conn = newClient.conn
	c.ch = newClient.ch
	c.queue = newClient.queue
	c.confirms = newClient.confirms
	return nil
}

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

	// ================================
	// 4. 宣告 Dead Letter Exchange / Queue
	// ================================
	// 4.1 宣告 DLX（死信交換機）
	if err := ch.ExchangeDeclare(
		cfg.DeadLetterExchange, // DLX 名稱，與 Consumer 端一致
		"direct",               // 類型
		true,                   // durable
		false,                  // auto-delete
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	); err != nil {
		return nil, fmt.Errorf("宣告 Dead Letter Exchange 失敗: %w", err)
	}

	// 4.2 宣告 DLQ（死信佇列）
	deadQ, err := ch.QueueDeclare(
		cfg.DeadLetterQueue, // DLQ 名稱，與 Consumer 端一致
		true,                // durable
		false,               // auto-delete
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("宣告 Dead Letter Queue 失敗: %w", err)
	}

	// 4.3 綁定 DLQ 到 DLX，routing key 為 "dead_ddldml"
	if err := ch.QueueBind(
		deadQ.Name,
		cfg.DeadLetterRoutingKey, // Dead Letter RoutingKey
		cfg.DeadLetterExchange,   // Dead Letter Exchange 名稱
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("綁定 Dead Letter Queue 失敗: %w", err)
	}

	// ================================
	// 5. 宣告主 Exchange（global_direct）
	// ================================
	if err := ch.ExchangeDeclare(
		cfg.PrimaryExchange, // 主交換機名稱
		"direct",            // 類型
		true,                // durable
		false,               // auto-delete
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	); err != nil {
		return nil, fmt.Errorf("宣告 Exchange 失敗: %w", err)
	}

	// =========================================
	// 6. 宣告主 Queue（ddl_dml_main_queue），並帶入 DLX 參數
	// =========================================
	DLXArgs := amqp.Table{
		"x-dead-letter-exchange":    cfg.DeadLetterExchange,   // 與 Consumer 相同
		"x-dead-letter-routing-key": cfg.DeadLetterRoutingKey, // 與 Consumer 相同
	}
	q, err := ch.QueueDeclare(
		cfg.PrimaryQueue, // queue 名稱
		true,             // durable
		false,            // auto-delete
		false,            // exclusive
		false,            // no-wait
		DLXArgs,          // 帶入 DLX 參數
	)
	if err != nil {
		return nil, fmt.Errorf("宣告 Primary Queue 失敗: %w", err)
	}

	// 7. Bind 多個 RoutingKey 到主 Queue
	//    （與 Consumer 端做法一致，只是把原本有參數 nil 的地方改成 DLXArgs）
	for _, routingKey := range AllRoutingKeys {
		if err := ch.QueueBind(
			q.Name,
			string(routingKey),  // 轉成 string 後對應 queue 中的 Binding
			cfg.PrimaryExchange, // 與 Consumer 端相同的 Exchange
			false,
			nil,
		); err != nil {
			return nil, fmt.Errorf("queue RoutingKey Bind(%s) 失敗：%v", string(routingKey), err)
		}
	}

	// 8. 回傳 MQClient 實例
	return &MQClient{conn: conn, ch: ch, cfg: cfg, queue: &q, confirms: confirms}, nil
}

// 發送訊息
func (c *MQClient) Publish(ctx context.Context, routingKey RoutingKey, body []byte) error {

	if err := c.ensureChannel(); err != nil {
		if err := c.Reconnect(); err != nil {
			return err
		}
		if err := c.ensureChannel(); err != nil {
			return err
		}
	}

	// 發送
	if err := c.ch.PublishWithContext(
		ctx,
		c.cfg.PrimaryExchange, // 改成自訂的 Exchange
		string(routingKey),    // routingKey = queue 名稱
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         body,
		}); err != nil {
		return err
	}

	// 等待 broker 回傳確認
	select {
	case confirm := <-c.confirms:
		if confirm.Ack {
			// 成功被 Queue 接收
			return nil
		}
		return fmt.Errorf("訊息被 broker Nack, Tag=%d", confirm.DeliveryTag)
	case <-time.After(c.cfg.ConfirmTimeout):
		return fmt.Errorf("publisher Confirm 超時")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *MQClient) Close() {
	c.ch.Close()
	c.conn.Close()
}
