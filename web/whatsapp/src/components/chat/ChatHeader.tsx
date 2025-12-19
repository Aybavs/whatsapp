import React from 'react';

import { Avatar } from '@/components/ui/avatar';
import { Contact, Group, User } from '@/types';

interface ChatHeaderProps {
  contact: Contact;
  isTyping?: boolean;
  className?: string;
  onSearchClick: () => void;
}

export const ChatHeader: React.FC<ChatHeaderProps> = ({
  contact,
  isTyping = false,
  className = "",
  onSearchClick,
}) => {
  const isGroup = (c: Contact): c is Group => {
      return (c as Group).is_group === true;
  };

  const displayName = isGroup(contact) ? contact.name : contact.name || contact.username;
  const avatarSrc = isGroup(contact) ? contact.avatar_url : (contact as User).avatar;
  
  // For group: show member count. For user: show status.
  const statusText = isGroup(contact) 
    ? `${contact.member_ids.length} members`
    : (contact as User).status === "online" ? "Online" : "Offline";

  const isOnline = !isGroup(contact) && (contact as User).status === "online";
  
  return (
    <div
      className={`flex items-center justify-between px-4 py-2 bg-white border-b ${className}`}
    >
      <div className="flex items-center">
        <Avatar
          name={displayName}
          size="md"
          src={avatarSrc}
        />
        <div className="ml-3 flex flex-col">
          <span className="font-medium text-sm">
            {displayName}
          </span>
          <div className="flex items-center gap-1 text-xs text-gray-500">
            {isTyping ? (
              <span className="text-green-600 animate-pulse">typing...</span>
            ) : (
              <>
                {isOnline && (
                  <span className="w-2 h-2 rounded-full bg-green-500 inline-block" />
                )}
                <span>{statusText}</span>
              </>
            )}
          </div>
        </div>
      </div>
      
      <button 
        onClick={onSearchClick}
        className="p-2 text-gray-500 hover:bg-gray-100 rounded-full transition-colors"
        title="Search messages"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="11" cy="11" r="8"/>
          <path d="m21 21-4.3-4.3"/>
        </svg>
      </button>
    </div>
  );
};
