import React, { useState, useContext } from "react";
import { bufferToBase64url, base64urlToBuffer } from "../utils/webauthn";
import { AuthContext } from "../contexts/AuthContext";

const LoginPage: React.FC = () => {
  const [username, setUsername] = useState("");
  const [message, setMessage] = useState("");
  const { refreshAuth } = useContext(AuthContext);

  const handleLogin = async () => {
    setMessage("Starting login...");

    try {
      // Step 1: Begin Login
      const beginResp = await fetch("http://localhost:8080/login/begin", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username }),
        credentials: "include",
      });

      if (!beginResp.ok) {
        const error = await beginResp.text();
        throw new Error(error);
      }

      const response = await beginResp.json();
      const publicKeyOptions = response.publicKey.publicKey;
      const userID = response.userID;

      // Convert options to proper format
      publicKeyOptions.challenge = base64urlToBuffer(
        publicKeyOptions.challenge,
      );
      if (publicKeyOptions.allowCredentials) {
        publicKeyOptions.allowCredentials =
          publicKeyOptions.allowCredentials.map(
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (cred: any) => ({
              id: base64urlToBuffer(cred.id),
              type: cred.type,
            }),
          );
      }

      // Add explicit authenticator selection options
      publicKeyOptions.authenticatorSelection = {
        requireResidentKey: false,
        userVerification: "preferred",
      };

      // Step 2: Get Assertion with timeout
      const assertion = (await navigator.credentials.get({
        publicKey: publicKeyOptions,
        signal: AbortSignal.timeout(60000),
      })) as PublicKeyCredential;

      // Step 3: Prepare Data to Send to Server
      const assertionData = {
        id: assertion.id,
        rawId: bufferToBase64url(assertion.rawId),
        type: assertion.type,
        response: {
          clientDataJSON: bufferToBase64url(assertion.response.clientDataJSON),
          authenticatorData: bufferToBase64url(
            (assertion.response as AuthenticatorAssertionResponse)
              .authenticatorData,
          ),
          signature: bufferToBase64url(
            (assertion.response as AuthenticatorAssertionResponse).signature,
          ),
          userHandle: bufferToBase64url(
            (assertion.response as AuthenticatorAssertionResponse).userHandle!,
          ),
        },
      };

      // Step 4: Finish Login
      const finishResp = await fetch(
        `http://localhost:8080/login/finish?userID=${encodeURIComponent(userID)}`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(assertionData),
          credentials: "include",
        },
      );

      if (!finishResp.ok) {
        const error = await finishResp.text();
        console.error("Detailed error:", error);
      }

      setMessage("Login successful!");
      refreshAuth();
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      console.error("Detailed error:", error);
      setMessage("Login failed: " + error.message);
    }
  };

  return (
    <div className="max-w-md mx-auto">
      <h1 className="text-2xl font-bold mb-4">Login</h1>
      <div className="mb-4">
        <label className="block">Username:</label>
        <input
          className="w-full border p-2"
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
      </div>
      <button
        className="bg-blue-500 text-white px-4 py-2"
        onClick={handleLogin}
      >
        Login
      </button>
      {message && <p className="mt-4">{message}</p>}
    </div>
  );
};

export default LoginPage;
