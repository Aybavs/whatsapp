import { AlertCircle, Loader2, SendIcon } from 'lucide-react';
import { memo, useEffect, useMemo, useRef } from 'react';

import { useAuth } from '@/hooks/useAuth';
import { Message } from '@/types';

import { MessageBubble } from './MessageBubble';

interface MessageListProps {
  messages: Message[];
  loading: boolean;
  error: string | null;
}

const MessageListComponent = ({ messages, loading, error }: MessageListProps) => {
  // Ensure messages is always an array
  const safeMessages = messages || [];
  const { user } = useAuth();
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Improved scroll logic
  useEffect(() => {
    if (safeMessages.length > 0 && containerRef.current) {
      // More reliable scroll to bottom
      setTimeout(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
      }, 100);
    }
  }, [safeMessages]);

  // Sort messages by date (oldest to newest) - memoized
  const sortedMessages = useMemo(() =>
    [...safeMessages].sort(
      (a, b) =>
        new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    ),
    [safeMessages]
  );

  // Group messages by date - memoized
  const groupedMessages = useMemo(() =>
    sortedMessages.reduce((groups, message) => {
      const date = new Date(message.created_at).toLocaleDateString();
      if (!groups[date]) groups[date] = [];
      groups[date].push(message);
      return groups;
    }, {} as Record<string, Message[]>),
    [sortedMessages]
  );

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-gray-50">
        <div className="flex flex-col items-center">
          <Loader2 className="h-8 w-8 text-green-500 animate-spin mb-2" />
          <p className="text-sm text-gray-500">Loading messages...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex-1 flex items-center justify-center bg-gray-50 p-4">
        <div className="flex flex-col items-center text-center">
          <AlertCircle className="h-8 w-8 text-red-500 mb-2" />
          <p className="text-sm font-medium text-red-500 mb-1">
            Unable to load messages
          </p>
          <p className="text-xs text-gray-500 max-w-xs">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className={`h-full w-full px-4 py-4 bg-opacity-60 bg-[url('/chat-bg-pattern.png')] bg-repeat ${
        safeMessages.length > 0 ? "overflow-y-auto" : "overflow-hidden"
      }`}
    >
      {safeMessages.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-full text-center text-gray-500 py-10">
          <div className="bg-white rounded-full p-4 shadow mb-4">
            <SendIcon className="h-8 w-8 text-green-500 opacity-50" />
          </div>
          <p className="text-sm font-medium mb-1">No messages yet</p>
          <p className="text-xs">
            Start the conversation by sending a message!
          </p>
        </div>
      ) : (
        <div className="min-h-full flex flex-col">
          <div className="space-y-4 mt-auto">
            {Object.entries(groupedMessages).map(([date, group]) => (
              <div key={date} className="space-y-2">
                <div className="flex justify-center">
                  <div className="bg-white px-3 py-1 rounded-full text-xs font-medium text-gray-500 shadow">
                    {date === new Date().toLocaleDateString() ? "Today" : date}
                  </div>
                </div>

                {group.map((message, index) => {
                  const isSameSenderAsPrevious =
                    index > 0 &&
                    group[index - 1].sender_id === message.sender_id;

                  return (
                    <MessageBubble
                      key={message.id}
                      message={message}
                      isSender={message.sender_id === user?.id}
                      showAvatar={!isSameSenderAsPrevious}
                    />
                  );
                })}
              </div>
            ))}
          </div>
        </div>
      )}

      <div ref={messagesEndRef} className="h-1" />
    </div>
  );
};

// MessageList'i memo ile sarmalayarak export et
export const MessageList = memo(MessageListComponent);
