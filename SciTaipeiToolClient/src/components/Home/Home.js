import React from "react";
import Menu from "../../utils/Menu";
// Styles migrated to Tailwind

const Home = ({ setToken }) => {
  return (
    <div className="font-sans p-5 bg-background min-h-screen">
      <Menu setToken={setToken} />
      <div className="max-w-2xl mx-auto bg-white p-5 rounded-lg shadow">
        <h2 className="text-xl font-bold mb-2">Home</h2>
        <p>Welcome to the home page.</p>
      </div>
    </div>
  );
};

export default Home;
