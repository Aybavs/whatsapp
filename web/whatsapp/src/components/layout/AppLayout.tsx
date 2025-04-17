import { ReactNode, useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';

import { AppHeader } from '@/components/layout/AppHeader';
import { useAuth } from '@/hooks/useAuth';
import { WebSocketStatus } from '@/services/WebSocketService/types';

interface AppLayoutProps {
  children: ReactNode;
  wsStatus: WebSocketStatus;
}

export const AppLayout: React.FC<AppLayoutProps> = ({ children, wsStatus }) => {
  const { user } = useAuth();
  const [isLoading, setIsLoading] = useState(true);
  const [stableStatus, setStableStatus] = useState<WebSocketStatus>(wsStatus);

  useEffect(() => {
    const timer = setTimeout(() => setIsLoading(false), 500);
    return () => clearTimeout(timer);
  }, []);

  useEffect(() => {
    if (wsStatus !== stableStatus) {
      const t = setTimeout(() => setStableStatus(wsStatus), 200);
      return () => clearTimeout(t);
    }
  }, [wsStatus, stableStatus]);

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
      <AppHeader stableStatus={stableStatus} user={user} />

      <main className="flex-1 flex flex-col md:flex-row w-full overflow-hidden text-gray-900">
        {children}
      </main>
    </div>
  );
};
