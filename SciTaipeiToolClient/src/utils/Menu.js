// Menu.js
import React from "react";
import { useNavigate } from "react-router-dom";
import apiClient from "./apiClient";

const Menu =  ({ setToken }) => {
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      // 調用後端 API 刪除 Refresh Token
      await apiClient.post("/users/Logout");

      // 清除 LocalStorage 的 Access Token
      localStorage.removeItem("accessToken");
      setToken(null); // 更新 token 為 null，觸發重定向
      alert("成功登出");
      // 跳轉到登入頁面
      
      // 跳轉到登入頁面
      navigate("/", { replace: true });
    } catch (error) {
      console.error("Logout failed:", error.response?.data?.message || error.message);
    }
  };

  const goHome = () => {
    navigate("/home");
  };

  const goScript = () => {
    navigate("/script");
  };

  const goServiceLog = () => {
    navigate("/service-log");
  };

  return (
    <nav style={styles.nav}>
      <h2 style={styles.logo}>My App</h2>
      <div style={styles.menu}>
      <button style={styles.menuButton} onClick={goHome}>
        Home
      </button>
      <button style={styles.menuButton} onClick={goScript}>
        Script
      </button>
      <button style={styles.menuButton} onClick={goServiceLog}>
        Service Log
      </button>
      <button style={styles.menuButton} onClick={handleLogout}>
        Logout
      </button>
      </div>
    </nav>
  );
};

const styles = {
  nav: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    backgroundColor: "#333",
    color: "white",
    padding: "10px 20px",
  },
  logo: {
    margin: 0,
  },
  menu: {
    display: "flex",
    gap: "10px",
  },
  menuButton: {
    backgroundColor: "#555",
    color: "white",
    border: "none",
    padding: "10px 15px",
    cursor: "pointer",
    borderRadius: "5px",
  },
};

export default Menu;
