import React, { useState } from "react";
import { Link,useNavigate } from "react-router-dom";
import {createApiClient} from "../../utils/apiClient"
// Styles migrated to Tailwind

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
    <div className="max-w-md mx-auto my-12 p-5 border border-gray-300 rounded-lg shadow bg-white text-center">
      <div className="mb-5">
        <h1 className="text-primary font-black text-2xl m-0">SCI Taipei Tool</h1>
      </div>
      <h2 className="mb-5 text-gray-800 font-bold">Forget Password</h2>
      <form onSubmit={handleSubmit}>
        <div className="mb-4 text-left">
          <label htmlFor="email" className="block mb-1 font-bold text-gray-800">Email:</label>
          <input
            type="email"
            id="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            className="w-full p-2 border border-gray-300 rounded box-border focus:border-primary focus:outline-none"
          />
        </div>
        <div className="mb-4 text-left">
          <label htmlFor="password" className="block mb-1 font-bold text-gray-800">New Password:</label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={handlePasswordChange}
            required
            className="w-full p-2 border border-gray-300 rounded box-border focus:border-primary focus:outline-none"
          />
        </div>
        <div className="mb-4 text-left">
          <label htmlFor="confirm-password" className="block mb-1 font-bold text-gray-800">Confirm New Password:</label>
          <input
            type="password"
            id="confirm-password"
            value={confirmPassword}
            onChange={handleConfirmPasswordChange}
            required
            className="w-full p-2 border border-gray-300 rounded box-border focus:border-primary focus:outline-none"
          />
          {!passwordMatch && (
            <p className="text-primary text-sm mt-1">Passwords do not match</p>
          )}
        </div>
        <button
          type="submit"
          className="w-full p-2 bg-primary text-white rounded text-lg cursor-pointer mt-2 disabled:bg-gray-300 disabled:text-gray-600 disabled:cursor-not-allowed hover:bg-primaryDark"
          disabled={!passwordMatch}
        >
          Reset Password
        </button>
      </form>
      <Link to="/" className="block mt-5 text-primary font-bold hover:underline">
        Back to Login
      </Link>
    </div>
  );
};

export default ForgetPassword;
