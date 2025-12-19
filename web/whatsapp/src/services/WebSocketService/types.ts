import { Message, TypingEvent } from '@/types';

export type WebSocketStatus = "connected" | "disconnected" | "connecting";

export type StatusUpdateEvent = {
  type: "status";
  UserID: string;
  status: string;
};
export type RawStatusUpdatePayload = {
  UserID: string;
  status: string;
};

export type MessageStatusUpdateEvent = {
  message_id: string;
  status: string;
  updated_at: string;
  sender_id?: string;
  receiver_id?: string;
};

export type BatchStatusUpdateEvent = {
  type: "batch";
  sender_id: string;
  receiver_id: string;
  status: string;
  updated_at: string;
};

export type MessageHandler = (data: Message) => void;
export type StatusUpdateHandler = (event: StatusUpdateEvent) => void;
export type TypingHandler = (event: TypingEvent) => void;
export type ErrorHandler = (error: Event) => void;
export type CloseHandler = (event: CloseEvent) => void;
export type StatusHandler = (status: WebSocketStatus) => void;
export type MessageStatusUpdateHandler = (event: MessageStatusUpdateEvent) => void;
export type BatchStatusUpdateHandler = (event: BatchStatusUpdateEvent) => void;

export interface EventHandlers {
  message: MessageHandler[];
  status: StatusHandler[];
  error: ErrorHandler[];
  close: CloseHandler[];
  statusUpdate: StatusUpdateHandler[];
  typing: TypingHandler[];
  messageStatusUpdate: MessageStatusUpdateHandler[];
  batchStatusUpdate: BatchStatusUpdateHandler[];
}
