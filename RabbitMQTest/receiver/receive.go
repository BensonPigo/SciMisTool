package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// 載入客戶端憑證與金鑰
	clientCert, err := tls.LoadX509KeyPair("certs/PmsApClient.crt", "certs/PmsApClient.key")
	failOnError(err, "Failed to load client certificate")

	// // 載入信任的 CA 憑證
	caCert, err := ioutil.ReadFile("certs/ca.crt")
	failOnError(err, "Failed to read CA certificate")

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// 建立 TLS 配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		// 若需要雙向驗證，則必須驗證伺服器的憑證
		RootCAs: caCertPool,
		// 若你的伺服器憑證使用正確的 CN 或 SAN，此處可設為 false
		InsecureSkipVerify: false,
	}

	// 使用 TLS 建立連線，連線字串改為 amqps 並連到 5671 端口
	conn, err := amqp.DialTLS("amqps://root:admin1234@localhost:5671/", tlsConfig)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// 宣告 Channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Exchange 宣告，這個Exange需要和生產者一樣
	err = ch.ExchangeDeclare(
		"logs",   // Exchange name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an Exchange")

	// Queue 1 宣告
	q, err := ch.QueueDeclare(
		"durable_queue1", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Queue 2 宣告
	q2, err := ch.QueueDeclare(
		"durable_queue2", // Queue name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// 將Exchange和兩個Queue binding
	err = ch.QueueBind(
		q.Name,   // queue name
		"Queue1", // routing key
		"logs",   // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind queue 1")
	err = ch.QueueBind(
		q2.Name,  // queue name
		"Queue2", // routing key
		"logs",   // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind queue 2")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	var forever chan struct{}

	// 監聽所有的Queue
	go func() {
		msgs, err := ch.Consume(
			q.Name, // 要監聽的 Queue 名稱
			"",     // consumer tag
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		failOnError(err, "Failed to register a consumer")

		for d := range msgs {
			log.Printf("Queue1 received a message: %s", d.Body)
			dotCount := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dotCount)
			time.Sleep(t * time.Second)
			log.Printf("Done")
			d.Ack(false) //手動回應
		}
	}()

	// 監聽所有的Queue
	go func() {
		msgs, err := ch.Consume(
			q2.Name, // 要監聽的 Queue 名稱
			"",      // consumer tag
			false,   // auto-ack
			false,   // exclusive
			false,   // no-local
			false,   // no-wait
			nil,     // args
		)
		failOnError(err, "Failed to register a consumer")

		for d := range msgs {
			log.Printf("Queue2 received a message: %s", d.Body)
			dotCount := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dotCount)
			time.Sleep(t * time.Second)
			log.Printf("Done")
			d.Ack(false) //手動回應
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
