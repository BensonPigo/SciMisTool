# SCI MIS Tool 前端指南

本專案為 **SCI MIS Tool** 的前端界面，使用 [React](https://react.dev/) 與 [Tailwind CSS](https://tailwindcss.com/) 製作。透過此界面可登入系統、執行遠端腳本並查詢服務記錄。

## 專案架構

- **React**：搭配 `react-router-dom` 管理路由。
- **Tailwind CSS**：負責整體樣式與版型。
- **axios**：`utils/apiClient.js` 內提供帶有攔截器的 API Client。

主要畫面組件位於 `src/components/`：

- `Login`：使用者登入。
- `Register`：註冊帳號。
- `ForgetPassword`：重設密碼。
- `Home`：登入後的首頁。
- `Script`：列出可執行的腳本並下達執行指令。
- `ServiceLog`：依條件查詢服務的 log。
- `Menu`：畫面上方的導覽列及登出功能。

## 環境變數

程式碼中使用以下環境變數，可在建置或啟動前設定於 `.env`：

- `REACT_APP_API_BASE_URL`：後端 API 的根網址，預設 `http://localhost:4795/api/v1`。
- `REACT_APP_FACTORY_IDS`：`ServiceLog` 頁面下拉選單的工廠編號，使用逗號分隔。
- `REACT_APP_SERVICE_NAMES`：`ServiceLog` 頁面服務名稱下拉選單，使用逗號分隔。

## 安裝與指令

確認已安裝 [Node.js](https://nodejs.org/) 與 [Yarn](https://yarnpkg.com/)，接著在 `SciTaipeiToolClient` 目錄下執行：

```bash
yarn install # 安裝依賴
```

開發模式：

```bash
yarn start
```

執行測試：

```bash
yarn test
```

建置生產版本：

```bash
yarn build
```

## 部署

`yarn build` 會在 `build/` 產生靜態檔案，可部署至任何支援靜態檔案的伺服器，例如 Nginx、Apache 或 GitHub Pages。亦可將此資料夾內容置於後端服務的靜態檔案路徑下供使用者瀏覽。

