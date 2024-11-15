import React, { useContext } from "react";
import { AuthContext } from "../contexts/AuthContext";

const ProtectedPage: React.FC = () => {
  const { isAuthenticated, user, loading } = useContext(AuthContext);

  if (loading) {
    return <p>Loading...</p>;
  }

  if (!isAuthenticated || !user) {
    return (
      <div>
        <h1 className="text-2xl font-bold mb-4">Access Denied</h1>
        <p>You are not logged in.</p>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Protected Page</h1>
      <p>Welcome, {user.displayName}!</p>
      <p>Your user ID is: {user.id}</p>
    </div>
  );
};

export default ProtectedPage;
