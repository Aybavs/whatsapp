import { useCallback, useEffect, useRef, useState } from 'react';

import WebSocketService from '@/services/WebSocketService';
import { TypingEvent } from '@/types';

export const useTypingIndicator = (contactId: string | null) => {
  const [isTyping, setIsTyping] = useState(false);
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const debounceRef = useRef<NodeJS.Timeout | null>(null);

  // Listen for incoming typing events
  useEffect(() => {
    if (!contactId) {
      setIsTyping(false);
      return;
    }

    const handleTypingEvent = (event: TypingEvent) => {
      if (event.sender_id === contactId) {
        setIsTyping(event.is_typing);

        // Auto-reset typing indicator after 3 seconds if no update
        if (event.is_typing) {
          if (typingTimeoutRef.current) {
            clearTimeout(typingTimeoutRef.current);
          }
          typingTimeoutRef.current = setTimeout(() => {
            setIsTyping(false);
          }, 3000);
        }
      }
    };

    const unsubscribe = WebSocketService.onTyping(handleTypingEvent);

    return () => {
      unsubscribe();
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
    };
  }, [contactId]);

  // Send typing event with debounce
  const sendTyping = useCallback((typing: boolean) => {
    if (!contactId) return;

    // Clear existing debounce
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    // Send typing event immediately
    WebSocketService.sendTypingEvent(contactId, typing);

    // If typing, auto-send false after 3 seconds of inactivity
    if (typing) {
      debounceRef.current = setTimeout(() => {
        WebSocketService.sendTypingEvent(contactId, false);
      }, 3000);
    }
  }, [contactId]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
    };
  }, []);

  return { isTyping, sendTyping };
};
