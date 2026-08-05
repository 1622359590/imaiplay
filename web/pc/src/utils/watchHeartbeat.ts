export interface WatchHeartbeatPayload {
  watched_seconds_delta: number;
  report_id: string;
}

export interface PlaybackSnapshot {
  positionSeconds: number;
  durationSeconds: number;
}

export interface PlaybackProgressReport {
  positionSeconds: number;
  progressPercent: number;
  heartbeat?: WatchHeartbeatPayload;
}

interface PlaybackLifecycleOptions {
  now: () => number;
  read: () => PlaybackSnapshot;
  report: (report: PlaybackProgressReport) => Promise<void>;
  reportIDFactory?: ReportIDFactory;
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

function progressReport(
  snapshot: PlaybackSnapshot,
  heartbeat?: WatchHeartbeatPayload,
): PlaybackProgressReport {
  const positionSeconds = Number.isFinite(snapshot.positionSeconds)
    ? Math.max(0, snapshot.positionSeconds)
    : 0;
  const progressPercent = Number.isFinite(snapshot.durationSeconds) && snapshot.durationSeconds > 0
    ? Math.max(0, Math.min(100, Math.floor((positionSeconds / snapshot.durationSeconds) * 100)))
    : 0;
  return {
    positionSeconds,
    progressPercent,
    ...(heartbeat ? { heartbeat } : {}),
  };
}

export class PlaybackLifecycleController {
  private readonly heartbeat: WatchHeartbeat;
  private mediaPlaying = false;
  private visible = true;
  private sampledAt: number | null = null;
  private failedReport: PlaybackProgressReport | null = null;

  constructor(private readonly options: PlaybackLifecycleOptions) {
    this.heartbeat = new WatchHeartbeat(options.reportIDFactory);
  }

  playing(): void {
    if (this.mediaPlaying) return;
    this.mediaPlaying = true;
    this.heartbeat.play();
    this.sampledAt = this.options.now();
  }

  async pause(): Promise<void> {
    this.stopPlayback();
    await this.flushHeartbeat();
  }

  async waiting(): Promise<void> {
    this.stopPlayback();
    await this.flushHeartbeat();
  }

  async ended(): Promise<void> {
    this.stopPlayback();
    const heartbeatReported = await this.flushHeartbeat();
    if (!heartbeatReported) await this.reportPositionOnly();
  }

  async visibilityChanged(visible: boolean): Promise<void> {
    if (visible === this.visible) return;
    if (!visible) {
      this.sample();
      this.visible = false;
      this.heartbeat.setVisible(false);
      this.sampledAt = null;
      await this.flushHeartbeat();
      return;
    }
    this.visible = true;
    this.heartbeat.setVisible(true);
    if (this.mediaPlaying) this.sampledAt = this.options.now();
  }

  async periodicFlush(): Promise<void> {
    this.sample();
    await this.flushHeartbeat();
  }

  async pagehide(): Promise<void> {
    this.sample();
    this.mediaPlaying = false;
    this.visible = false;
    this.heartbeat.pause();
    this.heartbeat.setVisible(false);
    this.sampledAt = null;
    await this.flushHeartbeat();
  }

  async seeked(): Promise<void> {
    await this.reportPositionOnly();
  }

  private async reportPositionOnly(): Promise<void> {
    try {
      await this.options.report(progressReport(this.options.read()));
    } catch {
      // Position-only compatibility updates remain non-blocking.
    }
  }

  private stopPlayback(): void {
    this.sample();
    this.mediaPlaying = false;
    this.heartbeat.pause();
    this.sampledAt = null;
  }

  private sample(): void {
    if (!this.mediaPlaying || !this.visible || this.sampledAt === null) return;
    const now = this.options.now();
    const elapsedMilliseconds = Math.max(0, now - this.sampledAt);
    this.sampledAt = now;
    this.heartbeat.addPlayedSeconds(elapsedMilliseconds / 1_000);
  }

  private async flushHeartbeat(): Promise<boolean> {
    const heartbeat = this.heartbeat.flush();
    if (!heartbeat) return false;
    const report = this.failedReport?.heartbeat?.report_id === heartbeat.report_id
      ? this.failedReport
      : progressReport(this.options.read(), heartbeat);
    try {
      await this.options.report(report);
      this.heartbeat.acknowledged(heartbeat.report_id);
      if (this.failedReport?.heartbeat?.report_id === heartbeat.report_id) {
        this.failedReport = null;
      }
    } catch {
      this.failedReport = report;
      this.heartbeat.failed(heartbeat);
    }
    return true;
  }
}
