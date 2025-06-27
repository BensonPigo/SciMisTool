# TpeBiConsumer

`TpeBiConsumer` 會從 RabbitMQ 監聽 DDL 與 DML 訊息並寫入 SQL Server。啟動時同樣會暴露 Prometheus Metrics 供監控使用。

## 功能概述

- 透過 TLS 連線至 RabbitMQ，依設定的 `consumer_count` 啟動多個 consumer
- 解析收到的訊息後執行 DDL 或 DML 變更
- 對執行過的記錄標記狀態避免重複處理
- 暴露批次次數、錯誤次數、處理耗時等 Prometheus 指標

## 專案結構

```
TpeBiConsumer/
├─ certs/        # TLS 憑證範例
├─ config/       # dev/prod 設定檔與初始化程式
├─ metrics/      # Prometheus 指標定義
├─ mq/           # RabbitMQ consumer 相關
├─ service/      # Processor 處理訊息並寫入 SQL Server
├─ model/        # MQ 訊息等資料結構
├─ scilog/       # 日誌工具
└─ main.go       # 程式進入點
```

程式啟動後會依 `consumer_count` 產生多個 consumer，同時監聽 DDL/DML
routing key，呼叫 `Processor` 完成資料庫更新。

## 建置與啟動

請指定 `dev` 或 `prod` build tag：

```bash
cd TpeBiConsumer
go build -tags dev -o tpe-bi-consumer
./tpe-bi-consumer
```

<!-- 亦可使用 `go run -tags dev .` 直接執行。Metrics 預設在 `:2113/metrics` 提供。 -->

### 設定覆寫

設定檔與環境變數規則與 Producer 相同，前綴為 `FtyBiProducer`。常見變數如下：

| 變數 | 說明 |
| ---- | ---- |
| `FtyBiProducer_MQ_AMQP_URL` | RabbitMQ 連線字串 |
| `FtyBiProducer_DB_HOST` | 資料庫主機 |
| `FtyBiProducer_DB_USER` | 資料庫帳號 |
| `FtyBiProducer_DB_PASSWORD` | 資料庫密碼 |

設定結構詳見 `config/config.go`。

## 發布與部署

正式環境建議以 `prod` build tag 建置：

```bash
go build -tags prod -o tpe-bi-consumer
```

將編譯後的可執行檔與 `config/config.prod.yaml`、`certs/` 目錄一併部署。
<!-- 啟動後服務會在 `:2113/metrics` 暴露監控指標，可配合 Prometheus 收集。 -->

