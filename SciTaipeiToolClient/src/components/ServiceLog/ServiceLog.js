import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import Menu from "../../utils/Menu";
import apiClient from "../../utils/apiClient";
import JsonGrid from "@redheadphone/react-json-grid";
// Styles migrated to Tailwind

const ServiceLog = ({ setToken }) => {
  const navigate = useNavigate();
  const [factoryId, setFactoryId] = useState("");
  const [serviceName, setServiceName] = useState("");
  const [logDate, setLogDate] = useState("");
  const [logData, setLogData] = useState(null);
  const [logSearchText, setlogSearchText] = useState("");
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
    <div className="font-sans p-5 bg-background min-h-screen">
      <Menu setToken={setToken} />
      <div className="max-w-2xl mx-auto bg-white p-5 rounded-lg shadow">
        <h2 className="text-xl font-bold mb-4">Service Log</h2>
        <div className="flex flex-wrap items-end gap-2 mb-4">
          <select
            value={factoryId}
            onChange={(e) => setFactoryId(e.target.value)}
            className="border border-gray-300 rounded p-2 flex-1"
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
            className="border border-gray-300 rounded p-2 flex-1"
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
            className="border border-gray-300 rounded p-2"
          />
          <button className="p-2 bg-primary text-white rounded hover:bg-primaryDark" onClick={handleQuery}>
            查詢
          </button>
        </div>
        {isLoading ? (
          <div className="text-center mt-12">
            <div className="w-10 h-10 border-4 border-gray-200 border-t-info rounded-full animate-spin mx-auto" />
          </div>
        ) : (
          logData && (
            <JsonGrid data={logData} enableSearch enableSorting />
          )
        )}
      </div>
    </div>
  );
};

export default ServiceLog;
