import { useCallback, useEffect, useState } from 'react';

import { messageApi } from '@/api/messageApi';
import { useAuth } from '@/hooks/useAuth';
import { useWebSocket } from '@/hooks/useWebSocket';
import WebSocketService from '@/services/WebSocketService';
import { Contact, Group, Message } from '@/types';

export const useChat = (selectedContact: Contact | null) => {
  const { user } = useAuth();
  useWebSocket();
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isGroup = selectedContact && (selectedContact as Group).is_group;

  // Fetch message history when contact changes
  useEffect(() => {
    const fetchMessages = async () => {
      if (!selectedContact || !user) return;

      setLoading(true);
      setError(null);
      try {
        const fetchedMessages = await messageApi.getMessages(
          selectedContact.id,
          !!isGroup
        );
        setMessages(fetchedMessages || []);
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
  }, [selectedContact, user, isGroup]);

  // Function to send a message
  const sendMessage = useCallback(
    async (content: string, mediaUrl?: string) => {
      if (!selectedContact || !user || (!content.trim() && !mediaUrl)) return false;

      const optimisticMessage: Message = {
        id: `optimistic-${Date.now()}`,
        sender_id: user.id,
        receiver_id: !isGroup ? selectedContact.id : '',
        content,
        media_url: mediaUrl,
        created_at: new Date().toISOString(),
        status: "sent",
      };

      if (isGroup) {
          (optimisticMessage as any).group_id = selectedContact.id;
      }

      setMessages((prev) => [...prev, optimisticMessage]);

      try {
        const newMessage = await messageApi.sendMessage(
          selectedContact.id,
          content,
          mediaUrl,
          !!isGroup
        );
        
        // Replace optimistic message with real one
        setMessages((prev) => 
            prev.map(m => m.id === optimisticMessage.id ? newMessage : m)
        );
        return true;
      } catch (err) {
        console.error("Failed to send message via API:", err);
        setError("Failed to send message");
        // Remove optimistic message on failure
         setMessages((prev) => prev.filter(m => m.id !== optimisticMessage.id));
        return false;
      }
    },
    [selectedContact, user, isGroup]
  );

  // Handle incoming WebSocket messages
  useEffect(() => {
    const handleWebSocketMessage = (message: Message & { group_id?: string }) => {
      if (!selectedContact) return;

      let shouldAdd = false;

      if (isGroup) {
        // Check if message belongs to this group
        if (message.group_id === selectedContact.id) {
          shouldAdd = true;
        }
      } else {
        // Direct message check
        if (
            (!message.group_id || message.group_id === "000000000000000000000000") && 
            (message.sender_id === selectedContact.id ||
             message.receiver_id === selectedContact.id)
        ) {
          shouldAdd = true;
        }
      }

      if (shouldAdd) {
        setMessages((prev) => {
          const currentMessages = prev || [];
          if (currentMessages.some((m) => m.id === message.id)) return currentMessages;
          return [...currentMessages, message];
        });

        if (!isGroup && message.sender_id === selectedContact.id && !message.id.startsWith("optimistic-") && !message.id.startsWith("temp-")) {
          messageApi
            .updateMessageStatus(message.id, "read")
            .catch((err) =>
              console.error("Failed to update message status:", err)
            );
        }
      }
    };

    const handleMessageStatusUpdate = (event: { message_id: string; status: string }) => {
        setMessages((prev) => 
            prev.map((msg) => 
                msg.id === event.message_id ? { ...msg, status: event.status as Message['status'] } : msg
            )
        );
    };

    const handleBatchStatusUpdate = (event: { sender_id: string; receiver_id: string; status: string }) => {
        if (selectedContact?.id === event.receiver_id) {
           setMessages((prev) =>
               prev.map((msg) =>
                   (msg.receiver_id === event.receiver_id && msg.status !== 'read')
                   ? { ...msg, status: event.status as Message['status'] }
                   : msg
               )
           );
        }
    };

    const unsubscribeMessage = WebSocketService.onMessage(handleWebSocketMessage);
    const unsubscribeStatus = WebSocketService.onMessageStatusUpdate(handleMessageStatusUpdate);
    const unsubscribeBatch = WebSocketService.onBatchStatusUpdate(handleBatchStatusUpdate);
    
    return () => {
        unsubscribeMessage();
        unsubscribeStatus();
        unsubscribeBatch();
    };
  }, [selectedContact, isGroup]);

  return {
    messages,
    loading,
    error,
    sendMessage,
    clearMessages: () => setMessages([]),
  };
};
