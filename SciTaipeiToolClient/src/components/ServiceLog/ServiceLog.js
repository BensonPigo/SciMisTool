import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import Menu from "../../utils/Menu";
import apiClient from "../../utils/apiClient";
import JsonGrid from "@redheadphone/react-json-grid";
import "./ServiceLog.css";

const ServiceLog = ({ setToken }) => {
  const navigate = useNavigate();
  const [factoryId, setFactoryId] = useState("");
  const [serviceName, setServiceName] = useState("");
  const [logDate, setLogDate] = useState("");
  const [logData, setLogData] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  const factoryOptions = (process.env.REACT_APP_FACTORY_IDS || "")
    .split(",")
    .filter((f) => f);
  const serviceOptions = (process.env.REACT_APP_SERVICE_NAMES || "")
    .split(",")
    .filter((s) => s);

  useEffect(() => {
    const token = localStorage.getItem("accessToken");
    if (!token) {
      setToken(null);
      navigate("/");
    }
  }, [setToken, navigate]);

  const handleQuery = async () => {
    if (!factoryId || !serviceName || !logDate) {
      alert("factoryId、serviceName、logDate皆為必填");
      return;
    }

    try {
      setIsLoading(true);
      const response = await apiClient.get("/service/log", {
        params: { factoryId, serviceName, logDate },
      });
      // API 回傳的 log 內容為 JSON 字串，需先轉成物件才能給 JsonGrid
      const rawData = response.data;
      const parsed =
        typeof rawData === "string" && rawData.trim() !== ""
          ? JSON.parse(rawData)
          : rawData;
      setLogData(parsed);
    } catch (error) {
      alert(error.response?.data.message || "Server 發生錯誤。");
      setLogData(null);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="home-container">
      <Menu setToken={setToken} />
      <div className="content">
        <h2>Service Log</h2>
        <div className="query-section">
          <select
            value={factoryId}
            onChange={(e) => setFactoryId(e.target.value)}
          >
            <option value="">Select Factory</option>
            {factoryOptions.map((f) => (
              <option key={f} value={f}>
                {f}
              </option>
            ))}
          </select>
          <select
            value={serviceName}
            onChange={(e) => setServiceName(e.target.value)}
          >
            <option value="">Select Service</option>
            {serviceOptions.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
          <input
            type="date"
            value={logDate}
            onChange={(e) => setLogDate(e.target.value)}
          />
          <button className="execute-btn" onClick={handleQuery}>
            查詢
          </button>
        </div>
        {isLoading ? (
          <div className="loading-container">
            <div className="spinner" />
          </div>
        ) : (
          logData && (
            <JsonGrid jsonData={logData} enableSearch enableSorting />
          )
        )}
      </div>
    </div>
  );
};

export default ServiceLog;
