# FtyBiProducer

`FtyBiProducer` 是一支以 Go 撰寫的批次服務，負責定期從 SQL Server 撈取 DDL/DML 紀錄並透過 RabbitMQ 發送至下游系統。程式亦內建 Prometheus Metrics，方便監控執行狀況。

## 主要功能

- 連接 SQL Server 取得尚未處理的 DDL 與 DML 紀錄
- 將紀錄封裝為 JSON，並以 TLS 方式連接 RabbitMQ 發送
- 支援訊息確認 (publisher confirm)
- 提供 Prometheus 指標：批次次數、錯誤次數與處理耗時

## 專案結構

```
FtyBiProducer/
├─ certs/        # 範例 TLS 憑證
├─ config/       # dev/prod 設定檔與初始化程式
├─ db/           # 資料庫存取函式
├─ metrics/      # Prometheus 指標定義
├─ mq/           # RabbitMQ 連線與發送
├─ service/      # Processor 邏輯，負責撈取並發送訊息
├─ model/        # DB 與 MQ 用到的資料結構
├─ scilog/       # 基於 zap 的簡易 logger
└─ main.go       # 程式進入點
```

`main.go` 會依序初始化設定、資料庫與 MQ 連線，接著啟動兩個批次處理
(`Processor.DdlLogProcess` 與 `Processor.DmlLogProcess`)，完成後將結果推送到 MQ。

## 建置方式

專案以 build tag 控制環境，請於建置或執行時指定 `dev` 或 `prod`：

```bash
# 於 dev 設定下建置
cd FtyBiProducer
go build -tags dev -o fty-bi-producer
```

若未指定 build tag 會導致編譯錯誤。各環境的預設設定檔位置如下：

- `dev`：`config/config.dev.yaml`
- `prod`：`config/config.prod.yaml`

## 執行

```bash
./fty-bi-producer
```

或直接以 `go run`：

```bash
go run -tags dev .
```

<!-- 程式啟動後會連線至資料庫與 MQ，並在 `:2112/metrics` 暴露 Prometheus 指標 (可於設定檔調整)。 -->

## 設定檔

設定使用 [viper](https://github.com/spf13/viper) 讀取，可透過環境變數覆寫，前綴為 `FtyBiProducer`。下列僅列出部分常用變數：

| 變數 | 說明 |
| ---- | ---- |
| `FtyBiProducer_MQ_AMQP_URL` | RabbitMQ 連線字串 |
| `FtyBiProducer_DB_HOST` | 資料庫主機 |
| `FtyBiProducer_DB_USER` | 資料庫帳號 |
| `FtyBiProducer_DB_PASSWORD` | 資料庫密碼 |

完整欄位請參考 `config/config.go` 內的結構定義。

## 發布與部署

正式環境建議以 `prod` build tag 編譯：

```bash
go build -tags prod -o fty-bi-producer
```

將 `fty-bi-producer`、`config/config.prod.yaml` 以及 `certs/` 目錄放置於同一
位置，必要時可透過環境變數覆寫設定。<!-- 啟動後服務會持續執行並暴露
`http://<host>:2112/metrics` 供監控使用。 -->

