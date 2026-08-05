export interface WatchHeartbeatPayload {
  watched_seconds_delta: number;
  report_id: string;
}

type ReportIDFactory = () => string;

const defaultReportIDFactory: ReportIDFactory = () => globalThis.crypto.randomUUID();

function samePayload(
  left: WatchHeartbeatPayload,
  right: WatchHeartbeatPayload,
): boolean {
  return left.report_id === right.report_id
    && left.watched_seconds_delta === right.watched_seconds_delta;
}

export class WatchHeartbeat {
  private playing = false;
  private visible = true;
  private accumulatedSeconds = 0;
  private inFlight: WatchHeartbeatPayload | null = null;
  private retry: WatchHeartbeatPayload | null = null;

  constructor(private readonly reportIDFactory: ReportIDFactory = defaultReportIDFactory) {}

  play(): void {
    this.playing = true;
  }

  pause(): void {
    this.playing = false;
  }

  setVisible(visible: boolean): void {
    this.visible = visible;
  }

  addPlayedSeconds(seconds: number): void {
    if (!this.playing || !this.visible || !Number.isFinite(seconds) || seconds <= 0) return;
    this.accumulatedSeconds += seconds;
  }

  flush(): WatchHeartbeatPayload | null {
    if (this.inFlight) return null;
    if (this.retry) {
      this.inFlight = this.retry;
      this.retry = null;
      return this.inFlight;
    }
    const watchedSeconds = Math.min(60, Math.floor(this.accumulatedSeconds));
    if (watchedSeconds < 1) return null;
    this.accumulatedSeconds -= watchedSeconds;
    this.inFlight = {
      watched_seconds_delta: watchedSeconds,
      report_id: this.reportIDFactory(),
    };
    return this.inFlight;
  }

  failed(payload: WatchHeartbeatPayload): void {
    if (!this.inFlight || !samePayload(this.inFlight, payload)) return;
    this.retry = this.inFlight;
    this.inFlight = null;
  }

  acknowledged(reportID: string): void {
    if (this.inFlight?.report_id === reportID) this.inFlight = null;
    if (this.retry?.report_id === reportID) this.retry = null;
  }
}
