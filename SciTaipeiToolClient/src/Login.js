import React, { useState } from "react";
import { Link ,useNavigate} from "react-router-dom";
import "./Login.css";
import { createApiClient } from "./apiClient";
import { saveAccessToken } from "./utils/token";

const Login =  ({ setToken }) => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const navigate = useNavigate(); 
  const apiWithoutInterceptor = createApiClient(false);

  const handleSubmit = async (e) => {
    e.preventDefault();

      try {
        const response = await apiWithoutInterceptor.post("/users/Login", {
          Email: email,
          Password: password,
        });

        if (response.status === 200 || response.status === 201) {

          // 儲存 Access Token
          saveAccessToken(response.data.accessToken);
          setToken(response.data.accessToken); // 更新 App 中的 token
          navigate("Home"); // Redirect to Login page
      }
      } catch (error) {
          if (error.response) {
              alert(error.response.data.message || "Something went wrong on the server.");
          } else {
              alert("Network error, please try again later.");
          }
      } finally {
      }
};

  return (
    <div className="login-container">
      <div className="logo-section">
        {/* 替換這段為 Logo 圖片或保持文字標題 */}
        <h1>SCI Taipei Tool</h1>
      </div>
      <h2>Login</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="email">Email:</label>
          <input
            type="email"
            id="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="password">Password:</label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        <button type="submit" className="login-button">
          Login
        </button>
      </form>
      <div className="additional-actions">
      <Link to="/register" className="action-button">
        Register
      </Link>
    <Link to="/forget-password" className="action-button">
      Forgot Password
    </Link>
    </div>
    </div>
  );
};

export default Login;
