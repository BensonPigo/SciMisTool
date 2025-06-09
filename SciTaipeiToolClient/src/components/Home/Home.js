import React from "react";
import Menu from "../../utils/Menu";
// Styles migrated to Tailwind

const Home = ({ setToken }) => {
  return (
    <div className="font-sans p-5 bg-background min-h-screen">
      <Menu setToken={setToken} />
      <div className="max-w-2xl mx-auto bg-white p-5 rounded-lg shadow">
        <h2 className="text-xl font-bold mb-2">Home</h2>
        <p className="mb-4">Welcome to the SCI MIS Tool. Use the menu above to access each feature.</p>
        <ul className="list-disc pl-5 space-y-1 text-gray-700">
          <li>Execute server scripts on the <strong>Script</strong> page.</li>
          <li>Query detailed logs on the <strong>Service Log</strong> page.</li>
        </ul>
      </div>
    </div>
  );
};

export default Home;
