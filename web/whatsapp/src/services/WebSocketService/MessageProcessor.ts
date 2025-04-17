import { Message, WebSocketMessage } from '@/types';

import { EventHandlerRegistry } from './EventHandlerRegistry';
import { RawStatusUpdatePayload, StatusUpdateEvent } from './types';

export class MessageProcessor {
  constructor(private handlers: EventHandlerRegistry) {}

  public handleRawMessage(data: string): void {
    if (data === "ping") {
      this.sendPong();
      return;
    }

    if (data === "pong") {
      this.handlers.triggerStatus("connected");
      return;
    }

    try {
      const parsed = JSON.parse(data);
      if (parsed.status && parsed.UserID) {
        this.handleStatusUpdate(parsed);
      } else {
        this.handleChatMessage(parsed);
      }
    } catch {
      // invalid JSON
    }
  }
  private transformMessageToInternalFormat(
    wsMessage: WebSocketMessage
  ): Message {
    return {
      id:
        wsMessage.id ||
        `temp-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
      sender_id: wsMessage.sender_id || "",
      receiver_id: wsMessage.receiver_id || "",
      content: wsMessage.content || "",
      media_url: wsMessage.media_url,
      created_at: wsMessage.created_at || new Date().toISOString(),
      status: wsMessage.status || "delivered",
    };
  }

  private handleStatusUpdate(data: RawStatusUpdatePayload): void {
    const event: StatusUpdateEvent = {
      type: "status",
      UserID: data.UserID,
      status: data.status,
    };
    this.handlers.triggerStatusUpdate(event);
  }

  private handleChatMessage(data: WebSocketMessage): void {
    const message: Message = this.transformMessageToInternalFormat(data);
    this.handlers.triggerMessage(message);
  }

  private sendPong(): void {
    const socket = this.getSocket();
    if (socket && socket.readyState === WebSocket.OPEN) {
      try {
        socket.send("pong");
      } catch {
        // Ignore errors when sending pong
      }
    }
  }

  private getSocket(): WebSocket | null {
    // Bu method dışarıdan socket alınacak şekilde genişletilebilir
    return null; // WebSocket dışarıdan enjekte edilmediyse null
  }
}
