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
    <nav className="flex justify-between items-center bg-gray-800 text-white px-4 py-2">
      <h2 className="m-0 font-semibold">SCI MIS Tool</h2>
      <div className="flex gap-2">
        <button
          className="bg-gray-600 text-white border-none px-3 py-2 rounded"
          onClick={goHome}
        >
          Home
        </button>
        <button
          className="bg-gray-600 text-white border-none px-3 py-2 rounded"
          onClick={goScript}
        >
          Script
        </button>
        <button
          className="bg-gray-600 text-white border-none px-3 py-2 rounded"
          onClick={goServiceLog}
        >
          Service Log
        </button>
        <button
          className="bg-gray-600 text-white border-none px-3 py-2 rounded"
          onClick={handleLogout}
        >
          Logout
        </button>
      </div>
    </nav>
  );
};

export default Menu;
