import {
  EventHandlers,
  MessageHandler,
  StatusHandler,
  ErrorHandler,
  CloseHandler,
  StatusUpdateHandler,
  StatusUpdateEvent,
  WebSocketStatus,
} from "./types";
import { Message } from "@/types";

export class EventHandlerRegistry {
  private handlers: EventHandlers = {
    message: [],
    status: [],
    error: [],
    close: [],
    statusUpdate: [],
  };

  public onMessage(handler: MessageHandler): () => void {
    this.handlers.message.push(handler);
    return () => {
      this.handlers.message = this.handlers.message.filter(
        (h) => h !== handler
      );
    };
  }

  public onStatus(handler: StatusHandler): () => void {
    this.handlers.status.push(handler);
    return () => {
      this.handlers.status = this.handlers.status.filter((h) => h !== handler);
    };
  }

  public onError(handler: ErrorHandler): () => void {
    this.handlers.error.push(handler);
    return () => {
      this.handlers.error = this.handlers.error.filter((h) => h !== handler);
    };
  }

  public onClose(handler: CloseHandler): () => void {
    this.handlers.close.push(handler);
    return () => {
      this.handlers.close = this.handlers.close.filter((h) => h !== handler);
    };
  }

  public onStatusUpdate(handler: StatusUpdateHandler): () => void {
    this.handlers.statusUpdate.push(handler);
    return () => {
      this.handlers.statusUpdate = this.handlers.statusUpdate.filter(
        (h) => h !== handler
      );
    };
  }

  public triggerStatus(status: WebSocketStatus): void {
    this.handlers.status.forEach((handler) => {
      try {
        handler(status);
      } catch {
        console.error("Error in status handler:", handler, status);
      }
    });
  }

  public triggerMessage(message: Message): void {
    this.handlers.message.forEach((handler) => handler(message));
  }

  public triggerStatusUpdate(event: StatusUpdateEvent): void {
    this.handlers.statusUpdate.forEach((handler) => handler(event));
  }

  public triggerError(error: Event): void {
    this.handlers.error.forEach((handler) => handler(error));
  }

  public triggerClose(event: CloseEvent): void {
    this.handlers.close.forEach((handler) => handler(event));
  }

  public removeStatusHandler(handler: StatusHandler): void {
    this.handlers.status = this.handlers.status.filter((h) => h !== handler);
  }

  public getHandlers(): EventHandlers {
    return this.handlers;
  }
}
