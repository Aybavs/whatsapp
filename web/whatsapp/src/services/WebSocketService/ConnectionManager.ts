export class ConnectionManager {
  private socket: WebSocket | null = null;

  public createSocket(url: string): WebSocket {
    this.cleanup();
    this.socket = new WebSocket(url);
    return this.socket;
  }

  public getSocket(): WebSocket | null {
    return this.socket;
  }

  public cleanup(): void {
    if (this.socket) {
      this.socket.onopen = null;
      this.socket.onmessage = null;
      this.socket.onerror = null;
      this.socket.onclose = null;
      this.socket.close();
      this.socket = null;
    }
  }

  public isConnectedOrConnecting(): boolean {
    return !!(
      this.socket &&
      (this.socket.readyState === WebSocket.CONNECTING ||
        this.socket.readyState === WebSocket.OPEN)
    );
  }

  public getSocketState(): number {
    return this.socket ? this.socket.readyState : -1;
  }

  public getSocketStateString(): string {
    if (!this.socket) return "null";
    const states = ["CONNECTING", "OPEN", "CLOSING", "CLOSED"];
    return states[this.socket.readyState] || "UNKNOWN";
  }
}
