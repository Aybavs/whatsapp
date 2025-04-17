import { Check, CheckCheck, Clock } from 'lucide-react';
import React from 'react';

import { cn } from '@/lib/utils';
import { Message } from '@/types';

interface MessageBubbleProps {
  message: Message;
  isSender: boolean;
  showAvatar?: boolean;
}

export const MessageBubble: React.FC<MessageBubbleProps> = ({
  message,
  isSender,
}) => {
  const formattedTime = new Date(message.created_at).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

  const renderStatusIcon = () => {
    if (!isSender) return null;

    switch (message.status) {
      case "sent":
        return <Check className="h-3 w-3 text-gray-400" />; // Tek tik
      case "delivered":
        return <CheckCheck className="h-3 w-3 text-gray-400" />; // Çift tik
      case "read":
        return <CheckCheck className="h-3 w-3 text-blue-500" />; // Mavi çift tik
      default:
        return <Clock className="h-3 w-3 text-gray-400" />; // Gönderilmekte
    }
  };

  // Mesaj durumuna göre arka plan rengini belirle
  const getBubbleBackground = () => {
    if (!isSender) return "bg-white";

    // Gönderen için durum bazlı arka plan renkleri
    switch (message.status) {
      case "sent":
      case "delivered":
      case "read":
        return "bg-[#dcf8c6]"; // Gönderildiğinde yeşil
      default:
        return "bg-[#f0f0f0]";
    }
  };

  return (
    <div
      className={cn(
        "flex mb-2 px-3 group",
        isSender ? "justify-end" : "justify-start"
      )}
    >
      <div
        className={cn(
          "relative max-w-[75%] px-4 py-2 rounded-2xl shadow-sm text-sm leading-relaxed whitespace-pre-wrap break-words",
          getBubbleBackground(),
          isSender
            ? "text-gray-900 rounded-tr-none"
            : "text-gray-800 rounded-tl-none border",
          message.status === undefined && isSender && "opacity-80"
        )}
      >
        {message.content}

        <div className="flex items-center justify-end gap-1 mt-1">
          <span className="text-[10px] text-gray-500">{formattedTime}</span>
          {renderStatusIcon()}
        </div>

        {/* Gönderilirken gösterge */}
        {!message.status && isSender && (
          <div className="absolute -bottom-4 right-1 text-[9px] text-gray-500 italic">
            sending...
          </div>
        )}
      </div>
    </div>
  );
};
