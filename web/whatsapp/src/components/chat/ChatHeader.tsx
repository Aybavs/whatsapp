import React from "react";
import { User } from "@/types";
import { Avatar } from "@/components/ui/avatar";

interface ChatHeaderProps {
  contact: User;
  className?: string;
}

export const ChatHeader: React.FC<ChatHeaderProps> = ({
  contact,
  className = "",
}) => {
  return (
    <div
      className={`flex items-center p-3 bg-gray-100 border-b border-gray-300 ${className}`}
    >
      <Avatar name={contact.name} />
      <div className="ml-3">
        <div className="font-medium">{contact.name}</div>
        <div className="text-xs text-gray-500">
          {contact.status || "offline"}
        </div>
      </div>
    </div>
  );
};
