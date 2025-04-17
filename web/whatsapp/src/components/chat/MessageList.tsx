import { useRef, useEffect } from "react";
import { useAuth } from "@/hooks/useAuth";
import { Message } from "@/types";
import { MessageBubble } from "./MessageBubble";

interface MessageListProps {
  messages: Message[];
  loading: boolean;
  error: string | null;
}

export const MessageList = ({ messages, loading, error }: MessageListProps) => {
  const { user } = useAuth();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="animate-pulse">Loading messages...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex-1 flex items-center justify-center text-red-500">
        Error loading messages: {error}
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-4">
      {messages.length === 0 ? (
        <div className="text-center text-gray-500 py-8">
          No messages yet. Start the conversation!
        </div>
      ) : (
        messages.map((message) => (
          <MessageBubble
            key={message.id}
            message={message}
            isSender={message.sender_id === user?.id}
          />
        ))
      )}
      <div ref={messagesEndRef} />
    </div>
  );
};