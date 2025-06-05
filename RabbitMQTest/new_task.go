package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		// data := map[string]string{
		// 	"M":                "PM2",
		// 	"FactoryID":        "MWI",
		// 	"Delivery":         "2025-08-11",
		// 	"Delivery(YYYYMM)": "202508",
		// }
		// b, err := json.Marshal(data)
		// if err != nil {
		// 	// 如果 JSON 產生有錯誤，則退回時間字串
		// 	s = time.Now().String()
		// } else {
		// 	s = string(b)
		// }

		s = time.Now().String()
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}

func NewTask() {

	// 載入客戶端憑證與私鑰
	clientCert, err := tls.LoadX509KeyPair("certs/PmsApClient.crt", "certs/PmsApClient.key")
	failOnError(err, "Failed to load client certificate")

	// 載入信任的 CA 憑證
	caCert, err := os.ReadFile("certs/ca.crt")
	failOnError(err, "Failed to read CA certificate")
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		log.Panic("Failed to append CA certificate")
	}

	// 建立 TLS 配置
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false, // 正式環境建議驗證伺服器憑證
	}

	// 使用 TLS 建立連線，連線字串改為 amqps 並連到 5671 端口
	conn, err := amqp.DialTLS("amqps://root:admin1234@localhost:5671/", tlsConfig)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Channel 宣告
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Exchange 宣告
	err = ch.ExchangeDeclare(
		"logs",   // Exchange name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// // Queue  宣告
	// q, err := ch.QueueDeclare(
	// 	"durable_queue1", // Queue name
	// 	true,             // durable
	// 	false,            // delete when unused
	// 	false,            // exclusive
	// 	false,            // no-wait
	// 	nil,              // arguments
	// )
	// failOnError(err, "Failed to declare a queue")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 準備要發送的data
	body := bodyFrom(os.Args)

	// 發送消息，可決定publish到exchange or queue
	ch.PublishWithContext(ctx,
		"logs", // exchange name
		// q.Name, // routing key
		"Queue2", // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", body)
}
