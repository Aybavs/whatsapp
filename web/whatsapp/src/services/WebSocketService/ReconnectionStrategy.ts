export class ReconnectionStrategy {
  private baseDelay = 1000;
  private attempts = 0;
  private maxAttempts = 3;
  private timer: NodeJS.Timeout | null = null;

  constructor(private reconnectFn: () => void) {}

  public handleClose(event: CloseEvent, isManual: boolean): void {
    const isAbnormal = event.code !== 1000 && event.code !== 1001;

    if (this.shouldReconnect(isAbnormal, isManual)) {
      this.scheduleReconnect();
    }
  }

  private shouldReconnect(isAbnormal: boolean, isManual: boolean): boolean {
    return isAbnormal && !isManual && this.attempts < this.maxAttempts;
  }

  private scheduleReconnect(): void {
    this.attempts++;
    const delay = Math.min(
      this.baseDelay * Math.pow(2, this.attempts - 1),
      30000
    );

    if (this.timer) {
      clearTimeout(this.timer);
    }

    this.timer = setTimeout(() => {
      this.reconnectFn();
    }, delay);
  }

  public reset(): void {
    this.attempts = 0;
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
  }

  public stop(): void {
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
  }
}
