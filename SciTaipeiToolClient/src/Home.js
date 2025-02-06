// Home.js
import React, { useState, useEffect } from "react";
import Menu from "./Menu";
import apiClient from "./apiClient";
import { useNavigate } from "react-router-dom";

const Home = ({ setToken }) => {
  const navigate = useNavigate();
  
  // 儲存從 API 取得的檔案路徑陣列
  const [files, setFiles] = useState([]);
  // 儲存使用者勾選的檔案（以完整路徑表示）
  const [selectedFiles, setSelectedFiles] = useState({});
   // 追蹤 Loading 狀態
  const [isLoading, setIsLoading] = useState(true);

  // 畫面載入時檢查 Token 與取得檔案資料
  useEffect(() => {
    const token = localStorage.getItem("accessToken");

    if (!token) {
      // Token 過期或不存在，清除並導回登入頁面
      setToken(null);
      navigate("/");
      return;
    }
    

    const fetchData = async () => {
        try {
          // 呼叫後端 API 取得檔案列表
          const response = await apiClient.get("/GetScripts");
    
          if (response.status === 200 || response.status === 201) {
            
            const results = response.data?.results || [];
                    
            if (Array.isArray(results)) {
                // 展開資料，將 FactoryId 和對應的 ScriptFiles 組合成一個平坦化的結構
                const flattenedFiles = results
                .filter(factory => factory.ScriptFiles.length > 0) 
                .map((factory) =>                    
                    factory.ScriptFiles.map((file) => ({
                    factoryId: factory.FactoryId,
                    fileName: file.split("\\").pop(), // 擷取檔案名稱
                    }))
                )
                .reduce((acc, curr) => acc.concat(curr), []); // 合併成單一陣列
        
                setFiles(flattenedFiles); // 儲存處理後的資料到狀態
            } else {
                console.error("Results is not an array:", results);
                setFiles([]); // 設置為空陣列，避免後續出錯
            }
          }
        } catch (error) {
          if (error.response) {
            alert(error.response.data.message || "Server 發生錯誤。");
          } else {
            alert(error);
          }
        } finally {
          setIsLoading(false); // 不管成功或失敗，都將 Loading 狀態設為 false
        }
      };

    fetchData();
    
}, [setToken, navigate]);


  // 處理 checkbox 狀態改變
  const handleCheckboxChange = (factoryId, fileName) => {
    setSelectedFiles((prevSelected) => {
      const currentFactoryFiles = prevSelected[factoryId] || [];
      const updatedFactoryFiles = currentFactoryFiles.includes(fileName)
        ? currentFactoryFiles.filter((file) => file !== fileName) // 移除選取
        : [...currentFactoryFiles, fileName]; // 新增選取

      return {
        ...prevSelected,
        [factoryId]: updatedFactoryFiles,
      };
    });
  };

  // 點擊 Execute 按鈕的處理邏輯
  const handleExecute = async () => {

    try {
        setIsLoading(true);

        const params = Object.entries(selectedFiles).map(([factoryId, scriptFiles]) => ({
            FactoryId: factoryId,
            TaskNames: scriptFiles
        }));
        const response = await apiClient.post("/ExecuteTask", params);
        
        if (response.status === 200 || response.status === 201) {
            alert(response.data.message);
        }
    } catch (error) {
      if (error.response) {
        alert(error.response.data.message || "Server 發生錯誤。");
      } else {
        alert(error);
      }
    } finally {
      setIsLoading(false); // 不管成功或失敗，都將 Loading 狀態設為 false
    }

    // console.log("選取的檔案：", selectedFiles);
    // const selectedFilesSummary = Object.entries(selectedFiles)
    //   .map(
    //     ([factoryId, files]) =>
    //       `${factoryId}: ${files.length > 0 ? files.join(", ") : "無選取檔案"}`
    //   )
    //   .join("\n");
    // alert(`選取的檔案清單：\n${selectedFilesSummary}`);
  };

  return (
    <div>
      <Menu setToken={setToken} />
      <div style={{ padding: "20px" }}>
        {/* <h1>Welcome to the Home Page</h1> */}

        {/* 顯示 Loading 動畫或檔案清單 */}
        {isLoading ? (
          <div style={{ textAlign: "center", marginTop: "20px" }}>
            <div className="spinner"></div>
            <p>Loading...</p>
          </div>
        ) : (
          <div>
            <h2>腳本清單</h2>
            <table border="1" cellPadding="5">
              <thead>
                <tr>
                  <th>選取</th>
                  <th>工廠Server</th>
                  <th>腳本名稱</th>
                </tr>
              </thead>
              <tbody>
                {files.length > 0 ? (
                  files.map((file, index) => (
                    <tr key={index}>
                      <td>
                        <input
                          type="checkbox"
                          checked={
                            selectedFiles[file.factoryId]?.includes(file.fileName) || false
                          }
                          onChange={() =>
                            handleCheckboxChange(file.factoryId, file.fileName)
                          }
                        />
                      </td>
                      <td>{file.factoryId}</td>
                      <td>{file.fileName}</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan="3">尚無檔案資料</td>
                  </tr>
                )}
              </tbody>
            </table>
            <br />
            <button onClick={handleExecute}>Execute</button>
          </div>
        )}
      </div>
    </div>
  );
};

export default Home;
