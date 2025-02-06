// 將 Access Token 存儲於 LocalStorage
export function saveAccessToken(token) {
    localStorage.setItem("accessToken", token);
  }
  
  // 從 LocalStorage 獲取 Access Token
  export function getAccessToken() {
    return localStorage.getItem("accessToken");
  }
  
  // 清除 LocalStorage 中的 Access Token
  export function clearAccessToken() {
    localStorage.removeItem("accessToken");
  }
  