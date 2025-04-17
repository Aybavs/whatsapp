import { useCallback, useEffect, useRef, useState } from 'react';

import webSocketService from '@/services/WebSocketService';
import {
    CloseHandler, ErrorHandler, MessageHandler, StatusUpdateHandler, WebSocketStatus
} from '@/services/WebSocketService/types';

import { useAuth } from './useAuth';

export const useWebSocket = () => {
  const { user, authLoaded } = useAuth();
  const [status, setStatus] = useState<WebSocketStatus>("disconnected");
  const connectCalled = useRef(false);

  const handleStatusChange = useCallback((newStatus: WebSocketStatus) => {
    setStatus(newStatus);
    connectCalled.current = newStatus === "connected";
  }, []);

  const attemptConnection = useCallback(() => {
    if (!user || connectCalled.current) return;

    connectCalled.current = true;
    webSocketService.onStatus(handleStatusChange);

    const connected = webSocketService.connect();
    if (!connected) connectCalled.current = false;

    // Ekstra: 2 saniye sonra socket açık mı kontrol et
    setTimeout(() => {
      const socket = webSocketService.getSocket();
      if (socket?.readyState === WebSocket.OPEN) {
        handleStatusChange("connected");
      }
    }, 2000);
  }, [user, handleStatusChange]);

  useEffect(() => {
    if (authLoaded && user && !connectCalled.current) {
      attemptConnection();
    }

    return () => {
      webSocketService.removeStatusHandler(handleStatusChange);
    };
  }, [authLoaded, user, attemptConnection, handleStatusChange]);

  const reconnect = useCallback(() => {
    webSocketService.disconnect();
    connectCalled.current = false;
    setStatus("disconnected");
    setTimeout(attemptConnection, 1000);
  }, [attemptConnection]);

  const getSocketDetails = useCallback(() => {
    const socket = webSocketService.getSocket();
    const socketState = socket?.readyState ?? -1;
    const stateMap = ["CONNECTING", "OPEN", "CLOSING", "CLOSED"];
    const socketStateString = stateMap[socketState] || "UNKNOWN";

    return {
      socketExists: !!socket,
      socketState,
      socketStateString,
    };
  }, []);

  return {
    status,
    reconnect,
    getSocketDetails,
    sendMessage: (receiverId: string, content: string, mediaUrl?: string) =>
      webSocketService.sendMessage(receiverId, content, mediaUrl),

    onMessage: (handler: MessageHandler) => webSocketService.onMessage(handler),
    onStatusUpdate: (handler: StatusUpdateHandler) =>
      webSocketService.onStatusUpdate(handler),
    onError: (handler: ErrorHandler) => webSocketService.onError(handler),
    onClose: (handler: CloseHandler) => webSocketService.onClose(handler),
  };
};
