export class HealthCheckManager {
  private checkInterval: NodeJS.Timeout | null = null;
  private pingInterval: NodeJS.Timeout | null = null;
  private lastPongTime = 0;

  constructor(
    private socketGetter: () => WebSocket | null,
    private onStale: () => void,
    private onAlive: () => void
  ) {}

  public start(): void {
    this.clear();
    this.checkInterval = setInterval(() => this.checkConnection(), 30000);
    this.pingInterval = setInterval(() => this.sendPing(), 30000);
  }

  public stop(): void {
    this.clear();
  }

  public updatePongTime(): void {
    this.lastPongTime = Date.now();
  }

  private clear(): void {
    if (this.checkInterval) {
      clearInterval(this.checkInterval);
      this.checkInterval = null;
    }
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  private checkConnection(): void {
    const socket = this.socketGetter();
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      this.onStale();
      return;
    }

    const now = Date.now();
    if (this.lastPongTime > 0 && now - this.lastPongTime > 60000) {
      this.onStale();
    } else {
      this.onAlive();
    }
  }

  private sendPing(): void {
    const socket = this.socketGetter();
    if (socket && socket.readyState === WebSocket.OPEN) {
      try {
        socket.send("ping");
      } catch {
        // Ignore errors when sending ping
      }
    }
  }
}
