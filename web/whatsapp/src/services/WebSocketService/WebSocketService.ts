import { ConnectionManager } from "./ConnectionManager";
import { EventHandlerRegistry } from "./EventHandlerRegistry";
import { HealthCheckManager } from "./HealthCheckManager";
import { MessageProcessor } from "./MessageProcessor";
import { ReconnectionStrategy } from "./ReconnectionStrategy";

export class WebSocketService {
  private static instance: WebSocketService;
  private connection = new ConnectionManager();
  private handlers = new EventHandlerRegistry();

  private health = new HealthCheckManager(
    () => this.connection.getSocket(),
    () => this.reconnect(),
    () => this.handlers.triggerStatus("connected")
  );

  private processor = new MessageProcessor(this.handlers);

  private reconnector = new ReconnectionStrategy(() => this.connect());

  private isManualDisconnect = false;

  private constructor() {}

  public static getInstance(): WebSocketService {
    if (!WebSocketService.instance) {
      WebSocketService.instance = new WebSocketService();
    }
    return WebSocketService.instance;
  }

  public connect(): boolean {
    if (this.connection.isConnectedOrConnecting()) {
      this.handlers.triggerStatus("connecting");
      return true;
    }

    const token = localStorage.getItem("token");
    if (!token) {
      this.handlers.triggerStatus("disconnected");
      return false;
    }

    const url = this.createUrl(token);
    const socket = this.connection.createSocket(url);

    socket.onopen = () => this.handleOpen();
    socket.onmessage = (e) => this.handleMessage(e);
    socket.onerror = (e) => this.handlers.triggerError(e);
    socket.onclose = (e) => this.handleClose(e);

    return true;
  }

  public disconnect(): void {
    this.isManualDisconnect = true;
    this.health.stop();
    this.reconnector.stop();
    this.connection.cleanup();
    this.handlers.triggerStatus("disconnected");
  }

  private handleOpen(): void {
    this.reconnector.reset();
    this.health.start();
    this.handlers.triggerStatus("connected");
  }

  private handleClose(event: CloseEvent): void {
    this.health.stop();
    this.handlers.triggerStatus("disconnected");
    this.handlers.triggerClose(event);
    this.reconnector.handleClose(event, this.isManualDisconnect);
  }

  private handleMessage(event: MessageEvent): void {
    if (event.data === "pong") {
      this.health.updatePongTime();
      return;
    }

    this.processor.handleRawMessage(event.data);
  }

  private reconnect(): void {
    this.connection.cleanup();
    this.connect();
  }

  private createUrl(token: string): string {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host; // Always use current host (e.g. localhost:3000)
    return `${protocol}//${host}/api/ws?token=${token}`;
  }

  public getSocket(): WebSocket | null {
    return this.connection.getSocket();
  }
  public sendMessage(
    receiverId: string,
    content: string,
    mediaUrl?: string
  ): boolean {
    const socket = this.connection.getSocket();
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      this.connect();
      return false;
    }

    const message = {
      receiver_id: receiverId,
      content,
      media_url: mediaUrl || null,
    };

    try {
      socket.send(JSON.stringify(message));
      return true;
    } catch {
      return false;
    }
  }

  public sendTypingEvent(receiverId: string, isTyping: boolean): boolean {
    const socket = this.connection.getSocket();
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return false;
    }

    const typingEvent = {
      type: "typing",
      receiver_id: receiverId,
      is_typing: isTyping,
    };

    try {
      socket.send(JSON.stringify(typingEvent));
      return true;
    } catch {
      return false;
    }
  }

  // Handler registration methods
  public onMessage = this.handlers.onMessage.bind(this.handlers);
  public onStatus = this.handlers.onStatus.bind(this.handlers);
  public onError = this.handlers.onError.bind(this.handlers);
  public onClose = this.handlers.onClose.bind(this.handlers);
  public onStatusUpdate = this.handlers.onStatusUpdate.bind(this.handlers);
  public onMessageStatusUpdate = this.handlers.onMessageStatusUpdate.bind(
    this.handlers
  );
  public onBatchStatusUpdate = this.handlers.onBatchStatusUpdate.bind(
    this.handlers
  );
  public onTyping = this.handlers.onTyping.bind(this.handlers);
  public removeStatusHandler = this.handlers.removeStatusHandler.bind(
    this.handlers
  );
}
