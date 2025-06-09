import React from "react";
import Menu from "../../utils/Menu";
import "./ServiceLog.css";
import JsonGrid from '@redheadphone/react-json-grid';

const ServiceLog = ({ setToken }) => {
  return (
    <div className="home-container">
      <Menu setToken={setToken} />
      <div className="content">
        <h2>Service Log</h2>
        <p>Coming soon.</p>
      </div>
    </div>
  );
};

export default ServiceLog;
