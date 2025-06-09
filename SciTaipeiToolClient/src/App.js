import React, { useState } from "react";
import "./App.css";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import Login from "./components/Login/Login";
import Register from "./components/Register/Register";
import ForgetPassword from "./components/ForgetPassword/ForgetPassword";
import Home from "./components/Home/Home";
import Script from "./components/Script/Script";
import ServiceLog from "./components/ServiceLog/ServiceLog";
import { getAccessToken } from "./utils/token";
import ProtectedRoute from "./utils/ProtectedRoute";

const App = () => {
  const [token, setToken] = useState(getAccessToken());

  return (
    <Router>
      <Routes>
        {/* 公開路由 */}
        <Route path="/register" element={<Register />} />
        <Route path="/forget-password" element={<ForgetPassword />} />

        {/* 登入路由 */}
        <Route
          path="/"
          element={token ? <Navigate to="/home" /> : <Login setToken={setToken} />}
        />

        {/* 受保護路由 */}
        <Route
          path="/home"
          element={
            <ProtectedRoute>
              <Home setToken={setToken} />
            </ProtectedRoute>
          }
        />
        <Route
          path="/script"
          element={
            <ProtectedRoute>
              <Script setToken={setToken} />
            </ProtectedRoute>
          }
        />
        <Route
          path="/service-log"
          element={
            <ProtectedRoute>
              <ServiceLog setToken={setToken} />
            </ProtectedRoute>
          }
        />
      </Routes>
    </Router>
  );
};

export default App;