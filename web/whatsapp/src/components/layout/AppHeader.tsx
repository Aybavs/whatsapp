import { FC } from 'react';

import { Avatar } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';
import { WebSocketStatus } from '@/services/WebSocketService/types';
import { User } from '@/types';

interface AppHeaderProps {
  stableStatus: WebSocketStatus;
  user: User;
}

export const AppHeader: FC<AppHeaderProps> = ({ stableStatus, user }) => {
  const { logout } = useAuth();

  return (
    <header className="bg-green-500 text-white px-6 py-4 shadow-sm">
      <div className="flex justify-between items-center max-w-7xl mx-auto">
        <h1 className="text-lg md:text-xl font-semibold tracking-wide">
          WhatsApp Clone
        </h1>

        <div className="flex items-center gap-5">
          <div className="flex items-center gap-2 text-sm">
            <div
              className={`h-2 w-2 rounded-full ${
                stableStatus === "connected" ? "bg-emerald-400" : "bg-red-400"
              }`}
            />
            <span className="capitalize text-white/90">{stableStatus}</span>
          </div>

          <div className="flex items-center gap-2">
            <Avatar name={user.name || user.username} size="sm" />
            <span className="text-sm font-medium">
              {user.name || user.username}
            </span>
          </div>

          <Button
            size="sm"
            onClick={logout}
            className="bg-white text-green-600 hover:bg-green-100 border border-white shadow-sm px-3 py-1 rounded text-sm"
          >
            Logout
          </Button>
        </div>
      </div>
    </header>
  );
};
