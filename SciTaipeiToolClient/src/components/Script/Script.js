// Script.js
import React, { useState, useEffect } from "react";
import Menu from "../../utils/Menu";
import apiClient from "../../utils/apiClient";
import { useNavigate } from "react-router-dom";
import DataTable from "react-data-table-component";
// Styles migrated to Tailwind

const Script = ({ setToken }) => {
  const navigate = useNavigate();
  const [files, setFiles] = useState([]);
  const [selectedFiles, setSelectedFiles] = useState({});
  const [isLoading, setIsLoading] = useState(true);
  const [filterText, setFilterText] = useState("");

  useEffect(() => {
    const token = localStorage.getItem("accessToken");
    if (!token) {
      setToken(null);
      navigate("/");
      return;
    }

    const fetchData = async () => {
      try {
        const response = await apiClient.get("/GetScripts");
        if (response.status === 200 || response.status === 201) {
          const results = response.data?.results || [];
          if (Array.isArray(results)) {
            const flattenedFiles = results
              .filter((factory) => factory.ScriptFiles.length > 0)
              .flatMap((factory) =>
                factory.ScriptFiles.map((file) => ({
                  factoryId: factory.FactoryId,
                  fileName: file.split("\\").pop(),
                }))
              );
            setFiles(flattenedFiles);
          } else {
            console.error("Results is not an array:", results);
            setFiles([]);
          }
        }
      } catch (error) {
        alert(error.response?.data.message || "Server 發生錯誤。");
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, [setToken, navigate]);

  const handleCheckboxChange = (factoryId, fileName) => {
    setSelectedFiles((prevSelected) => {
      const currentFactoryFiles = prevSelected[factoryId] || [];
      const updatedFactoryFiles = currentFactoryFiles.includes(fileName)
        ? currentFactoryFiles.filter((file) => file !== fileName)
        : [...currentFactoryFiles, fileName];
      return { ...prevSelected, [factoryId]: updatedFactoryFiles };
    });
  };

  const handleExecute = async () => {
    try {
      setIsLoading(true);
      const params = Object.entries(selectedFiles).map(([factoryId, scriptFiles]) => ({
        FactoryId: factoryId,
        TaskNames: scriptFiles,
      }));
      const response = await apiClient.post("/ExecuteTask", params);
      if (response.status === 200 || response.status === 201) {
        alert(response.data.message);
      }
    } catch (error) {
      alert(error.response?.data.message || "Server 發生錯誤。");
    } finally {
      setIsLoading(false);
    }
  };

  const filteredFiles = files.filter(
    (file) =>
      file.factoryId.toLowerCase().includes(filterText.toLowerCase()) ||
      file.fileName.toLowerCase().includes(filterText.toLowerCase())
  );

  const columns = [
    {
      name: "選取",
      cell: (row) => (
        <input
          type="checkbox"
          checked={
            selectedFiles[row.factoryId]?.includes(row.fileName) || false
          }
          onChange={() => handleCheckboxChange(row.factoryId, row.fileName)}
        />
      ),
      width: "100px",
    },
    {
      name: "工廠Server",
      selector: (row) => row.factoryId,
      sortable: true,
    },
    {
      name: "腳本名稱",
      selector: (row) => row.fileName,
      sortable: true,
    },
  ];

  return (
    <div className="font-sans p-5 bg-background min-h-screen">
      <Menu setToken={setToken} />
      <div className="max-w-2xl mx-auto bg-white p-5 rounded-lg shadow">
        <div className="bg-yellow-100 text-yellow-800 p-4 border border-yellow-200 rounded mb-5 text-sm">
          <p>請根據工作排程名稱，選擇對應的腳本，然後點選 "Execute"。<br />
          注意：這將直接執行正式 Server 的工作排程，請務必確認後再執行。</p>
        </div>
        <input
          type="text"
          placeholder="輸入關鍵字過濾資料..."
          value={filterText}
          onChange={(e) => setFilterText(e.target.value)}
          className="w-full p-2 mb-5 border border-gray-300 rounded box-border focus:border-info focus:outline-none focus:ring-2 focus:ring-info/50"
        />
        {isLoading ? (
          <div className="text-center mt-12">
            <div className="w-10 h-10 border-4 border-gray-200 border-t-info rounded-full animate-spin mx-auto" />
            <p>Loading...</p>
          </div>
        ) : (
          <div className="table-container">
            <h2 className="text-lg font-bold mb-2">腳本清單</h2>
            <DataTable
              columns={columns}
              data={filteredFiles}
              pagination
              highlightOnHover
              responsive
            />
            <button className="w-full p-2 bg-primary text-white rounded text-lg cursor-pointer mt-4 hover:bg-primaryDark" onClick={handleExecute}>Execute</button>
          </div>
        )}
      </div>
    </div>
  );
};

export default Script;
