import React, { useState } from "react";
import "./App.css";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import Login from "./Login";
import Register from "./Register";
import ForgetPassword from "./ForgetPassword";
import Home from "./Home";
import { getAccessToken } from "./utils/token";
import ProtectedRoute from "./ProtectedRoute";

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
      </Routes>
    </Router>
  );
};

export default App;