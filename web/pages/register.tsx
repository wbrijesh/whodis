import React, { useState } from "react";
import { bufferToBase64url, base64urlToBuffer } from "../utils/webauthn";

const RegisterPage: React.FC = () => {
  const [username, setUsername] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [message, setMessage] = useState("");

  const handleRegister = async () => {
    setMessage("Starting registration...");

    try {
      // Step 1: Begin Registration
      const beginResp = await fetch("http://localhost:8080/register/begin", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, displayName }),
        credentials: "include",
      });

      if (!beginResp.ok) {
        const errorText = await beginResp.text();
        try {
          const errorJson = JSON.parse(errorText);
          throw new Error(errorJson.message || errorText);
        } catch {
          throw new Error(errorText);
        }
      }

      const response = await beginResp.json();
      const publicKeyOptions = response.publicKey.publicKey;
      const userID = response.userID;

      // Convert options to proper format
      publicKeyOptions.user.id = base64urlToBuffer(publicKeyOptions.user.id);
      publicKeyOptions.challenge = base64urlToBuffer(
        publicKeyOptions.challenge,
      );
      publicKeyOptions.excludeCredentials = (
        publicKeyOptions.excludeCredentials || []
      )
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        .map((cred: any) => ({
          id: base64urlToBuffer(cred.id),
          type: cred.type,
        }));

      // Step 2: Create Credential
      const credential = (await navigator.credentials.create({
        publicKey: publicKeyOptions,
      })) as PublicKeyCredential;

      // Step 3: Prepare Data to Send to Server
      const credentialData = {
        id: credential.id,
        rawId: bufferToBase64url(credential.rawId),
        type: credential.type,
        response: {
          clientDataJSON: bufferToBase64url(credential.response.clientDataJSON),
          attestationObject: bufferToBase64url(
            (credential.response as AuthenticatorAttestationResponse)
              .attestationObject,
          ),
        },
      };

      // Step 4: Finish Registration
      const finishResp = await fetch(
        `http://localhost:8080/register/finish?userID=${encodeURIComponent(userID)}`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(credentialData),
          credentials: "include",
        },
      );

      if (!finishResp.ok) {
        const errorText = await finishResp.text();
        try {
          const errorJson = JSON.parse(errorText);
          throw new Error(errorJson.message || errorText);
        } catch {
          throw new Error(errorText);
        }
      }

      setMessage("Registration successful!");
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      console.error(error);
      setMessage("Registration failed: " + error.message);
    }
  };

  return (
    <div className="max-w-md mx-auto">
      <h1 className="text-2xl font-bold mb-4">Register</h1>
      <div className="mb-4">
        <label className="block">Username:</label>
        <input
          className="w-full border p-2"
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
      </div>
      <div className="mb-4">
        <label className="block">Display Name:</label>
        <input
          className="w-full border p-2"
          type="text"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
        />
      </div>
      <button
        className="bg-blue-500 text-white px-4 py-2"
        onClick={handleRegister}
      >
        Register
      </button>
      {message && <p className="mt-4">{message}</p>}
    </div>
  );
};

export default RegisterPage;
