import React from 'react';

import { Avatar } from '@/components/ui/avatar';
import { User } from '@/types';

interface ChatHeaderProps {
  contact: User;
  className?: string;
}

export const ChatHeader: React.FC<ChatHeaderProps> = ({
  contact,
  className = "",
}) => {
  const isOnline = contact.status?.toLowerCase() === "online";

  return (
    <div
      className={`flex items-center px-4 py-2 bg-white border-b ${className}`}
    >
      <Avatar
        name={contact.name || contact.username}
        size="md"
        src={contact.avatar}
      />
      <div className="ml-3 flex flex-col">
        <span className="font-medium text-sm">
          {contact.name || contact.username}
        </span>
        <div className="flex items-center gap-1 text-xs text-gray-500">
          {isOnline && (
            <span className="w-2 h-2 rounded-full bg-green-500 inline-block" />
          )}
          <span>{isOnline ? "Online" : "Offline"}</span>
        </div>
      </div>
    </div>
  );
};
