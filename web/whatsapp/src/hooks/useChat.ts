import { useCallback, useEffect, useState } from 'react';

import { messageApi } from '@/api/messageApi';
import { useAuth } from '@/hooks/useAuth';
import { useWebSocket } from '@/hooks/useWebSocket';
import WebSocketService from '@/services/WebSocketService';
import { Message, User } from '@/types';

export const useChat = (selectedContact: User | null) => {
  const { user } = useAuth();
  const { status: wsStatus, sendMessage: wsSendMessage } = useWebSocket();
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch message history when contact changes
  useEffect(() => {
    const fetchMessages = async () => {
      if (!selectedContact || !user) return;

      setLoading(true);
      setError(null);
      try {
        const fetchedMessages = await messageApi.getMessages(
          selectedContact.id
        );
        setMessages(fetchedMessages);
      } catch (err) {
        console.error("Failed to fetch messages:", err);
        setError("Failed to load message history");
      } finally {
        setLoading(false);
      }
    };

    if (selectedContact) {
      fetchMessages();
    } else {
      setMessages([]);
    }
  }, [selectedContact, user]);

  // Function to send a message
  const sendMessage = useCallback(
    async (content: string, mediaUrl?: string) => {
      if (!selectedContact || !user || !content.trim()) return false;

      if (wsStatus === "connected") {
        const sent = wsSendMessage(selectedContact.id, content, mediaUrl);
        if (sent) {
          const optimisticMessage: Message = {
            id: `optimistic-${Date.now()}`,
            sender_id: user.id,
            receiver_id: selectedContact.id,
            content,
            media_url: mediaUrl,
            created_at: new Date().toISOString(),
            status: "sent",
          };
          setMessages((prev) => [...prev, optimisticMessage]);

          // API'ye arka planda mesajı kaydet
          messageApi
            .sendMessage(selectedContact.id, content, mediaUrl)
            .catch((err) => {
              console.error("Failed to persist message via API:", err);
              setError("Message not saved");
            });

          return true;
        }
      }

      // WebSocket başarısızsa fallback olarak HTTP kullan
      try {
        const newMessage = await messageApi.sendMessage(
          selectedContact.id,
          content,
          mediaUrl
        );
        setMessages((prev) => [...prev, newMessage]);
        return true;
      } catch (err) {
        console.error("Failed to send message via API:", err);
        setError("Failed to send message");
        return false;
      }
    },
    [selectedContact, user, wsStatus, wsSendMessage]
  );

  // Handle incoming WebSocket messages
  useEffect(() => {
    const handleWebSocketMessage = (message: Message) => {
      // Only add messages from/to the selected contact
      if (
        selectedContact &&
        (message.sender_id === selectedContact.id ||
          message.receiver_id === selectedContact.id)
      ) {
        // Don't add duplicate messages
        setMessages((prev) => {
          if (prev.some((m) => m.id === message.id)) return prev;
          return [...prev, message];
        });

        // If message is from the selected contact, mark as read
        if (message.sender_id === selectedContact.id) {
          messageApi
            .updateMessageStatus(message.id, "read")
            .catch((err) =>
              console.error("Failed to update message status:", err)
            );
        }
      }
    };

    const unsubscribe = WebSocketService.onMessage(handleWebSocketMessage);
    return unsubscribe;
  }, [selectedContact]);

  return {
    messages,
    loading,
    error,
    sendMessage,
    clearMessages: () => setMessages([]),
  };
};
