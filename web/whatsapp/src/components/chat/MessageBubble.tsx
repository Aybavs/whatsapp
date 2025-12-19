import { Check, CheckCheck, Clock } from 'lucide-react';
import React, { memo } from 'react';

import { cn } from '@/lib/utils';
import { Message } from '@/types';
import { getAvatarColor } from '@/utils/colors';

interface MessageBubbleProps {
  message: Message;
  isSender: boolean;
  showAvatar?: boolean;
}

const MessageBubbleComponent: React.FC<MessageBubbleProps> = ({
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
        {/* Sender Name for Groups */}
        {!isSender && message.sender_username && (
          <div 
            className="text-xs font-bold mb-1"
            style={{ color: getAvatarColor(message.sender_username) }}
          >
            {message.sender_username}
          </div>
        )}

        {/* Media Content */}
        {message.media_url && (
          <div className="mb-2">
            {message.media_url.match(/\.(jpg|jpeg|png|gif|webp)$/i) ? (
              <img 
                src={message.media_url} 
                alt="Shared media" 
                className="rounded-lg max-w-full max-h-60 object-cover cursor-pointer hover:opacity-90 transition-opacity"
                onClick={() => window.open(message.media_url, '_blank')}
              />
            ) : (
              <a 
                href={message.media_url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="flex items-center gap-2 bg-black/5 p-2 rounded-lg hover:bg-black/10 transition-colors"
              >
                <div className="bg-gray-200 p-2 rounded">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
                  </svg>
                </div>
                <span className="text-sm underline truncate max-w-[150px]">
                  {message.media_url.split('/').pop()}
                </span>
              </a>
            )}
          </div>
        )}

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

// Custom comparison: sadece message.id, message.status veya isSender değişirse re-render
export const MessageBubble = memo(MessageBubbleComponent, (prevProps, nextProps) => {
  return (
    prevProps.message.id === nextProps.message.id &&
    prevProps.message.status === nextProps.message.status &&
    prevProps.isSender === nextProps.isSender
  );
});
