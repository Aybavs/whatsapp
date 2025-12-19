import { Message, TypingEvent, WebSocketMessage } from '@/types';

import { EventHandlerRegistry } from './EventHandlerRegistry';
import { RawStatusUpdatePayload, StatusUpdateEvent, MessageStatusUpdateEvent, BatchStatusUpdateEvent } from './types';

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

      console.log("WS Received:", parsed);

      // Handle typing events
      if (parsed.type === "typing") {
        this.handleTypingEvent(parsed);
        return;
      }

      // Handle batch status update (e.g. Read Receipts for multiple messages)
      if (parsed.type === "batch") {
          this.handleBatchStatusUpdate(parsed);
          return;
      }
      
      // Handle Message Status Update (Read/Delivered)
      if (parsed.message_id && parsed.status) {
         // This is a message status update, NOT a new message
         // We should trigger a specific handler for this
         // But currently EventHandlerRegistry might not have a specific trigger for message status
         // Let's check EventHandlerRegistry.ts next.
         // For now, I'll assume we need to add a triggerMessageStatus update
         
         // WAIT, I need to check if EventHandlerRegistry supports this. 
         // If not, I should probably reuse handleStatusUpdate but make sure the type is correct?
         // Or better, let's verify what 'triggerStatusUpdate' does.
         // 'handleStatusUpdate' above creates { type: "status", UserID: data.UserID, status: data.status }
         // This seems specific to USER status (online/offline).
         
         // I should probably start by checking EventHandlerRegistry.
         // But for this step, I will add a new method to handleMessageStatusUpdate
         // and ensure handleChatMessage is ONLY called if it's actually a message (has content).
         
         this.handleMessageStatusUpdate(parsed);
         return;
      }

      // Handle User Status Update (Online/Offline)
      if (parsed.status && parsed.UserID) {
        this.handleStatusUpdate(parsed);
        return;
      }
      
      // Handle Regular Chat Message
      // It must have content to be a message
      if (parsed.content) {
        this.handleChatMessage(parsed);
        return;
      }
      
      console.warn("Unknown message format:", parsed);

    } catch (e) {
      console.error("Error parsing WS message:", e);
    }
  }

  private handleMessageStatusUpdate(data: any): void {
      const event: MessageStatusUpdateEvent = {
          message_id: data.message_id,
          status: data.status,
          updated_at: data.updated_at,
          sender_id: data.sender_id,
          receiver_id: data.receiver_id
      };
      this.handlers.triggerMessageStatus(event);
  }

  private handleBatchStatusUpdate(data: any): void {
      const event: BatchStatusUpdateEvent = {
          type: "batch",
          sender_id: data.sender_id,
          receiver_id: data.receiver_id,
          status: data.status,
          updated_at: data.updated_at
      };
      this.handlers.triggerBatchStatus(event);
  }

  private handleTypingEvent(data: TypingEvent): void {
    this.handlers.triggerTyping(data);
  }
  private transformMessageToInternalFormat(
    wsMessage: WebSocketMessage
  ): Message {
    return {
      id:
        wsMessage.id ||
        `temp-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
      sender_id: wsMessage.sender_id || "",
      sender_username: wsMessage.sender_username,
      receiver_id: wsMessage.receiver_id || "",
      group_id: wsMessage.group_id,
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
