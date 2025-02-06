import React from "react";
import { Navigate } from "react-router-dom";

const ProtectedRoute = ({ children }) => {
  const token = localStorage.getItem("accessToken");

  if (!token) {
    // alert("Token 已失效，請重新登入");
    return <Navigate to="/" replace />;
  }

  return children;
};

export default ProtectedRoute;
