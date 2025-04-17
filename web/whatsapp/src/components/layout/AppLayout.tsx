import { ReactNode, useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';

import { Avatar } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';
import { useWebSocket } from '@/hooks/useWebSocket';
import { WebSocketStatus } from '@/services/WebSocketService/types';
import { User } from '@/types';

interface AppLayoutProps {
  children: ReactNode;
  wsStatus: WebSocketStatus;
}

interface DebugInfoProps {
  wsStatus: WebSocketStatus;
  lastStatusTime: Date;
  user: User;
  socketDetails: {
    socketExists: boolean;
    socketState: number;
    socketStateString: string;
  };
  showExtra?: boolean;
}

const DebugInfo: React.FC<DebugInfoProps> = ({
  wsStatus,
  lastStatusTime,
  user,
  socketDetails,
  showExtra = false,
}) => {
  return (
    <div className="mt-4 p-4 bg-white border border-gray-200 rounded-md text-xs text-gray-900 shadow-sm">
      <h3 className="text-sm font-semibold mb-2">Debug Info</h3>
      <p>
        <strong>Status:</strong> {wsStatus}
      </p>
      <p>
        <strong>Last Status Change:</strong>{" "}
        {lastStatusTime.toLocaleTimeString()}
      </p>
      <p>
        <strong>User:</strong> {user.name || user.username} (ID:{" "}
        {user.id?.substring(0, 8)}...)
      </p>
      <p>
        <strong>Token Present:</strong>{" "}
        {localStorage.getItem("token") ? "Yes" : "No"}
      </p>
      <p className="mt-2 font-bold">Socket Details:</p>
      <ul className="list-disc list-inside space-y-1">
        <li>
          <strong>Socket Exists:</strong>{" "}
          {socketDetails.socketExists ? "Yes" : "No"}
        </li>
        <li>
          <strong>Socket State:</strong> {socketDetails.socketStateString} (
          {socketDetails.socketState})
        </li>
        {showExtra && (
          <>
            <li>
              <strong>URL:</strong> {window.location.href}
            </li>
            <li>
              <strong>WebSocket URL:</strong>{" "}
              {window.location.protocol === "https:" ? "wss:" : "ws:"}
              //localhost:8080/api/ws?token=***
            </li>
          </>
        )}
      </ul>
    </div>
  );
};

export const AppLayout: React.FC<AppLayoutProps> = ({ children, wsStatus }) => {
  const { user, logout } = useAuth();
  const { reconnect, getSocketDetails } = useWebSocket();
  const [showDebugPanel, setShowDebugPanel] = useState(false);
  const [lastStatus, setLastStatus] = useState<WebSocketStatus>(wsStatus);
  const [lastStatusTime, setLastStatusTime] = useState(new Date());
  const [isLoading, setIsLoading] = useState(true);
  const [stableStatus, setStableStatus] = useState<WebSocketStatus>(wsStatus);
  const [socketDetails, setSocketDetails] = useState(getSocketDetails());

  useEffect(() => {
    const initialLoadTimer = setTimeout(() => setIsLoading(false), 500);
    return () => clearTimeout(initialLoadTimer);
  }, []);

  useEffect(() => {
    if (wsStatus !== lastStatus) {
      setLastStatus(wsStatus);
      setLastStatusTime(new Date());
      setSocketDetails(getSocketDetails());
      if (wsStatus === "connected") {
        setStableStatus("connected");
      } else {
        const timer = setTimeout(() => setStableStatus(wsStatus), 200);
        return () => clearTimeout(timer);
      }
    }
  }, [wsStatus, lastStatus, getSocketDetails]);

  if (!user) return <Navigate to="/login" />;

  if (isLoading) {
    return (
      <div className="flex flex-col min-h-screen items-center justify-center bg-white">
        <div className="animate-spin rounded-full h-14 w-14 border-4 border-whatsapp-green border-t-transparent"></div>
        <p className="mt-4 text-gray-900">Loading...</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col min-h-screen bg-white">
      <header className="bg-green-500 text-white px-4 py-3 shadow">
        <div className="flex justify-between items-center">
          <h1 className="text-xl font-bold tracking-wide">WhatsApp Clone</h1>
          <div className="flex items-center gap-4">
            <div className="flex items-center">
              <div
                className={`h-2 w-2 rounded-full mr-2 ${
                  stableStatus === "connected" ? "bg-green-400" : "bg-red-400"
                }`}
              />
              <span className="text-sm capitalize text-white/90">
                {stableStatus}
              </span>
            </div>
            <div className="flex items-center space-x-2">
              <Avatar name={user.name || user.username} size="sm" />
              <span className="text-sm font-medium text-white">
                {user.name || user.username}
              </span>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={logout}
              className="border-white text-white bg-transparent hover:bg-white/10"
            >
              Logout
            </Button>
          </div>
        </div>
      </header>

      <main className="flex-1 flex flex-col md:flex-row w-full overflow-hidden text-gray-900">
        {children}
      </main>

      <div className="fixed bottom-2 right-2">
        <button
          onClick={() => setShowDebugPanel((prev) => !prev)}
          className="bg-gray-200 text-gray-900 text-xs px-2 py-1 rounded shadow hover:bg-gray-300 transition"
        >
          {showDebugPanel ? "Hide Debug" : "Show Debug"}
        </button>
      </div>

      {showDebugPanel && (
        <div className="fixed bottom-12 right-2 bg-white p-4 rounded shadow-lg text-xs max-w-xs z-50 border border-gray-200">
          <DebugInfo
            wsStatus={wsStatus}
            lastStatusTime={lastStatusTime}
            user={user}
            socketDetails={socketDetails}
            showExtra
          />
          <div className="mt-3 flex gap-2 justify-end">
            <Button
              size="sm"
              className="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded text-xs"
              onClick={reconnect}
            >
              Force Reconnect
            </Button>
            <Button
              size="sm"
              variant="outline"
              className="border-gray-300 text-gray-800 hover:bg-gray-100 px-3 py-1 rounded text-xs"
              onClick={() => window.location.reload()}
            >
              Reload
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};
