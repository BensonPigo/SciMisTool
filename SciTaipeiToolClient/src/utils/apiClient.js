import axios from "axios";
import { getAccessToken, saveAccessToken, clearAccessToken } from "./token.js";

const BASE_URL = process.env.REACT_APP_API_BASE_URL || "http://localhost:4795/api/v1";

console.log("當前環境：", process.env.NODE_ENV);

/**
 * 創建 API Client，可選擇是否使用攔截器
 * @param {boolean} useInterceptor 是否使用攔截器，預設 true
 * @returns {AxiosInstance} axios 實例
 */
export function createApiClient(useInterceptor = true) {
  const apiClient = axios.create({
    baseURL: BASE_URL,
  });

  if (useInterceptor) {
    // 添加請求攔截器
    apiClient.interceptors.request.use(
      (config) => {
        const token = getAccessToken();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // 添加響應攔截器
    apiClient.interceptors.response.use(
      (response) => response,
      async (error) => {
        if (error.response?.status === 401) {
          try {
            // 嘗試刷新 Access Token
            const refreshResponse = await axios.post(`${BASE_URL}/users/RefreshToken`);
            const newAccessToken = refreshResponse.data.accessToken;

            // 更新 Access Token 並重試請求
            saveAccessToken(newAccessToken);
            error.config.headers.Authorization = `Bearer ${newAccessToken}`;
            return apiClient(error.config); // 重試原請求
          } catch (refreshError) {
            console.error("Refresh Token failed:", refreshError.response?.data?.message || refreshError.message);

            // 清除 Access Token 並重定向到登入頁
            clearAccessToken();
            window.location.href = "/login";
            return Promise.reject(refreshError);
          }
        }
        return Promise.reject(error);
      }
    );
  }

  return apiClient;
}

// 預設導出一個帶有攔截器的 API Client
const apiClient = createApiClient();
export default apiClient;
