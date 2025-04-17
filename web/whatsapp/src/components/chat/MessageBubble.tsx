import React from "react";
import { Message } from "@/types";

interface MessageBubbleProps {
  message: Message;
  isSender: boolean;
}

export const MessageBubble: React.FC<MessageBubbleProps> = ({
  message,
  isSender,
}) => {
  const formattedTime = new Date(message.created_at).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

  return (
    <div
      className={`max-w-[70%] mb-2 rounded-lg px-3 py-2 ${
        isSender ? "ml-auto bg-whatsapp-green bg-opacity-25" : "bg-white"
      }`}
    >
      <div className="text-sm">{message.content}</div>
      <div className="flex justify-end items-center mt-1 text-[10px] text-gray-500">
        <span>{formattedTime}</span>

        {isSender && (
          <span className="ml-1">
            {message.status === "delivered"
              ? "✓"
              : message.status === "read"
              ? "✓✓"
              : ""}
          </span>
        )}
      </div>
    </div>
  );
};
