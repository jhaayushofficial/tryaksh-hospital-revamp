import { initializeApp } from "firebase/app";
import { getAuth, type Auth } from "firebase/auth";
import { logger } from "./logger";

/**
 * Firebase is used for one thing only: proving the patient controls the phone
 * number they typed. Everything after that runs against the Tryaksh API.
 *
 * Initialization is lazy and guarded: `getAuth` throws synchronously on a bad
 * API key, and doing that at module load would crash the entire site —
 * including the doctor list and every other page that never touches phone
 * verification — before React even mounts. Deferring it to first use means a
 * Firebase misconfiguration only breaks the one step that needs it.
 */
const firebaseConfig = {
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
  storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET,
  messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGING_SENDER_ID,
  appId: import.meta.env.VITE_FIREBASE_APP_ID,
};

let auth: Auth | null = null;
let initError: Error | null = null;

function init(): Auth {
  if (auth) return auth;
  if (initError) throw initError;

  const missing = Object.entries(firebaseConfig)
    .filter(([, value]) => !value)
    .map(([key]) => key);

  if (missing.length > 0) {
    initError = new Error("Phone verification isn't configured on this deployment.");
    logger.error("firebase config is incomplete — phone verification will fail", { missing });
    throw initError;
  }

  try {
    auth = getAuth(initializeApp(firebaseConfig));
    return auth;
  } catch (cause) {
    initError = new Error("Couldn't start phone verification. Try again shortly.");
    logger.error("firebase initialization failed", { detail: String(cause) });
    throw initError;
  }
}

/** Returns the Auth instance, initializing it on first call. */
export function getFirebaseAuth(): Auth {
  return init();
}
