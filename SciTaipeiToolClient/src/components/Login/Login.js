import React, { useState } from "react";
import { Link ,useNavigate} from "react-router-dom";
// Styles migrated to Tailwind
import { createApiClient } from "../../utils/apiClient";
import { saveAccessToken } from "../../utils/token";

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
          navigate("/home"); // Redirect to Home page
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
        {/* 替換這段為 Logo 圖片或保持文字標題 */}
        <h1 className="text-primary font-black text-2xl m-0">SCI MIS  Tool</h1>
      </div>
      <h2 className="mb-5 text-gray-800 font-bold">Login</h2>
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
          <label htmlFor="password" className="block mb-1 font-bold text-gray-800">Password:</label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            className="w-full p-2 border border-gray-300 rounded box-border focus:border-primary focus:outline-none"
          />
        </div>
        <button type="submit" className="w-full p-2 bg-primary text-white rounded text-lg cursor-pointer mt-2 hover:bg-primaryDark">
          Login
        </button>
      </form>
      <div className="flex justify-between gap-2 mt-4">
        <Link to="/register" className="flex-1 p-2 bg-gray-800 text-white rounded text-center text-sm hover:bg-gray-700">
          Register
        </Link>
        <Link to="/forget-password" className="flex-1 p-2 bg-gray-800 text-white rounded text-center text-sm hover:bg-gray-700">
          Forgot Password
        </Link>
      </div>
    </div>
  );
};

export default Login;
