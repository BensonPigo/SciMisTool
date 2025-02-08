import React, { useState } from "react";
import { Link,useNavigate } from "react-router-dom";
import { createApiClient } from "./apiClient";
import "./ForgetPassword.css"; // 引用樣式

const ForgetPassword = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordMatch, setPasswordMatch] = useState(true);
  const navigate = useNavigate();
  const apiWithoutInterceptor = createApiClient(false);
  
  const handlePasswordChange = (e) => {
    const value = e.target.value;
    setPassword(value);
    setPasswordMatch(value === confirmPassword);
  };

  const handleConfirmPasswordChange = (e) => {
    const value = e.target.value;
    setConfirmPassword(value);
    setPasswordMatch(value === password);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!passwordMatch) {
      alert("Passwords do not match. Please try again.");
      return;
    }

    try {
        const response = await apiWithoutInterceptor.patch("/users/ResetPassword", {
            Email: email,
            Password: password,
        });

        if (response.status === 200 || response.status === 204) {
            alert(response.data.message);
            navigate("/"); // Redirect to Login page
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
    <div className="forget-password-container">
      <div className="logo-section">
        <h1>SCI Taipei Tool</h1>
      </div>
      <h2>Forget Password</h2>
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
          <label htmlFor="password">New Password:</label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={handlePasswordChange}
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="confirm-password">Confirm New Password:</label>
          <input
            type="password"
            id="confirm-password"
            value={confirmPassword}
            onChange={handleConfirmPasswordChange}
            required
          />
          {!passwordMatch && (
            <p className="error-message">Passwords do not match</p>
          )}
        </div>
        <button
          type="submit"
          className="reset-button"
          disabled={!passwordMatch}
        >
          Reset Password
        </button>
      </form>
    <Link to="/" className="back-to-login">
    Back to Login
    </Link>
    </div>
  );
};

export default ForgetPassword;
