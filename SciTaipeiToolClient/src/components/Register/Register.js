import React, { useState } from "react";
import {Link,useNavigate} from "react-router-dom"
import "./Register.css";
import { createApiClient } from "../../utils/apiClient";

const Register = () => {
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordMatch, setPasswordMatch] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const navigate = useNavigate();
  const apiWithoutInterceptor = createApiClient(false);

  const handlePasswordChange = (e) => {
    const value = e.target.value;
    setPassword(value);
    // 即時檢查是否一致
    setPasswordMatch(value === confirmPassword);
  };

  const handleConfirmPasswordChange = (e) => {
    const value = e.target.value;
    setConfirmPassword(value);
    // 即時檢查是否一致
    setPasswordMatch(value === password);
  };


const handleSubmit = async (e) => {
    e.preventDefault();

    if (!username || !email || !password) {
        alert("All fields are required.");
        return;
    }

    if (password !== confirmPassword) {
        alert("Passwords do not match. Please try again.");
        return;
    }

    setIsSubmitting(true);

    try {
        const response = await apiWithoutInterceptor.post("/users/Register", {
            Email: email,
            Username: username,
            Password: password,
        });

        if (response.status === 200 || response.status === 201) {
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
        setIsSubmitting(false);
    }
};


  return (
    <div className="register-container">
      <div className="logo-section">
        <h1>SCI Taipei Tool</h1>
      </div>
      <h2>Register</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="username">Username:</label>
          <input
            type="text"
            id="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
        </div>
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
            onChange={handlePasswordChange}
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="password">Confirm Password:</label>
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
          className="register-button"
          disabled={!passwordMatch || isSubmitting} // 禁用提交按鈕，直到密碼一致
        >
            {isSubmitting ? "Submitting..." : "Register"}
        </button>
        
        {/* <button 
        type="submit" 
        className="register-button" 
        disabled={isSubmitting}>
            {isSubmitting ? "Submitting..." : "Register"}
</button> */}
      </form>
    <Link to="/" className="back-to-login">
    Back to Login
    </Link>
    </div>
  );
};

export default Register;
