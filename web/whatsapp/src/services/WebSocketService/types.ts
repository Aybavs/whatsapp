import { Message } from '@/types';

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

export type MessageHandler = (data: Message) => void;
export type StatusUpdateHandler = (event: StatusUpdateEvent) => void;
export type ErrorHandler = (error: Event) => void;
export type CloseHandler = (event: CloseEvent) => void;
export type StatusHandler = (status: WebSocketStatus) => void;

export interface EventHandlers {
  message: MessageHandler[];
  status: StatusHandler[];
  error: ErrorHandler[];
  close: CloseHandler[];
  statusUpdate: StatusUpdateHandler[];
}
