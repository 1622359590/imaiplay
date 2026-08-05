import { describe, expect, it } from 'vitest';
import * as heartbeatModule from './watchHeartbeat';
import { WatchHeartbeat } from './watchHeartbeat';

interface PlaybackSnapshot {
  positionSeconds: number;
  durationSeconds: number;
}

interface PlaybackReport {
  positionSeconds: number;
  progressPercent: number;
  heartbeat?: {
    watched_seconds_delta: number;
    report_id: string;
  };
}

interface PlaybackLifecycle {
  playing(): void;
  pause(): Promise<void>;
  waiting(): Promise<void>;
  ended(): Promise<void>;
  visibilityChanged(visible: boolean): Promise<void>;
  periodicFlush(): Promise<void>;
  pagehide(): Promise<void>;
  pageshow(): Promise<void>;
  seeked(): Promise<void>;
}

type PlaybackLifecycleConstructor = new (options: {
  now: () => number;
  read: () => PlaybackSnapshot;
  report: (report: PlaybackReport) => Promise<void>;
  terminalReport?: (report: PlaybackReport) => Promise<void> | void;
  reportIDFactory: () => string;
}) => PlaybackLifecycle;

const PlaybackLifecycleController = (
  heartbeatModule as unknown as { PlaybackLifecycleController: PlaybackLifecycleConstructor }
).PlaybackLifecycleController;

describe('WatchHeartbeat', () => {
  it('accumulates only while playback is active and the document is visible', () => {
    const heartbeat = new WatchHeartbeat(() => 'report-1');
    heartbeat.addPlayedSeconds(4);
    heartbeat.play();
    heartbeat.addPlayedSeconds(15);
    heartbeat.pause();
    heartbeat.addPlayedSeconds(5);
    heartbeat.play();
    heartbeat.setVisible(false);
    heartbeat.addPlayedSeconds(6);

    expect(heartbeat.flush()).toEqual({
      watched_seconds_delta: 15,
      report_id: 'report-1',
    });
  });

  it('caps a report at 60 seconds and retains the remainder after acknowledgement', () => {
    let id = 0;
    const heartbeat = new WatchHeartbeat(() => `report-${++id}`);
    heartbeat.play();
    heartbeat.addPlayedSeconds(75);

    const first = heartbeat.flush();
    expect(first).toEqual({ watched_seconds_delta: 60, report_id: 'report-1' });
    expect(heartbeat.flush()).toBeNull();
    heartbeat.acknowledged('report-1');
    expect(heartbeat.flush()).toEqual({ watched_seconds_delta: 15, report_id: 'report-2' });
  });

  it('retries the identical payload and only clears a matching acknowledgement', () => {
    const heartbeat = new WatchHeartbeat(() => 'stable-report');
    heartbeat.play();
    heartbeat.addPlayedSeconds(15);
    const first = heartbeat.flush();
    expect(first).not.toBeNull();

    heartbeat.failed(first!);
    expect(heartbeat.flush()).toBe(first);
    heartbeat.acknowledged('different-report');
    expect(heartbeat.flush()).toBeNull();
    heartbeat.acknowledged('stable-report');
    expect(heartbeat.flush()).toBeNull();
  });

  it('does not send sub-second, invalid, or non-positive samples', () => {
    const heartbeat = new WatchHeartbeat(() => 'report-1');
    heartbeat.play();
    heartbeat.addPlayedSeconds(0.5);
    heartbeat.addPlayedSeconds(-1);
    heartbeat.addPlayedSeconds(Number.NaN);
    expect(heartbeat.flush()).toBeNull();
    heartbeat.addPlayedSeconds(0.5);
    expect(heartbeat.flush()).toEqual({ watched_seconds_delta: 1, report_id: 'report-1' });
  });
});

describe('PlaybackLifecycleController', () => {
  it('flushes only wall-clock seconds played while the page is visible', async () => {
    let now = 0;
    let snapshot = { positionSeconds: 0, durationSeconds: 60 };
    const reports: PlaybackReport[] = [];
    let nextID = 0;
    const controller = new PlaybackLifecycleController({
      now: () => now,
      read: () => snapshot,
      report: async (report) => { reports.push(report); },
      reportIDFactory: () => `report-${++nextID}`,
    });

    controller.playing();
    now = 5_000;
    snapshot = { positionSeconds: 5, durationSeconds: 60 };
    await controller.pause();

    now = 20_000;
    await controller.periodicFlush();

    controller.playing();
    now = 24_000;
    snapshot = { positionSeconds: 9, durationSeconds: 60 };
    await controller.visibilityChanged(false);
    now = 30_000;
    await controller.visibilityChanged(true);
    now = 33_000;
    snapshot = { positionSeconds: 12, durationSeconds: 60 };
    await controller.waiting();

    controller.playing();
    now = 35_000;
    snapshot = { positionSeconds: 60, durationSeconds: 60 };
    await controller.ended();

    controller.playing();
    now = 36_000;
    await controller.pagehide();

    expect(reports).toEqual([
      { positionSeconds: 5, progressPercent: 8, heartbeat: { watched_seconds_delta: 5, report_id: 'report-1' } },
      { positionSeconds: 9, progressPercent: 15, heartbeat: { watched_seconds_delta: 4, report_id: 'report-2' } },
      { positionSeconds: 12, progressPercent: 20, heartbeat: { watched_seconds_delta: 3, report_id: 'report-3' } },
      { positionSeconds: 60, progressPercent: 100, heartbeat: { watched_seconds_delta: 2, report_id: 'report-4' } },
      { positionSeconds: 60, progressPercent: 100, heartbeat: { watched_seconds_delta: 1, report_id: 'report-5' } },
    ]);
  });

  it('flushes accumulated play time on the deterministic 15-second periodic sample', async () => {
    let now = 0;
    const reports: PlaybackReport[] = [];
    const controller = new PlaybackLifecycleController({
      now: () => now,
      read: () => ({ positionSeconds: 15, durationSeconds: 100 }),
      report: async (report) => { reports.push(report); },
      reportIDFactory: () => 'periodic-report',
    });

    controller.playing();
    now = 10_000;
    controller.playing();
    now = 15_000;
    await controller.periodicFlush();

    expect(reports).toEqual([{
      positionSeconds: 15,
      progressPercent: 15,
      heartbeat: { watched_seconds_delta: 15, report_id: 'periodic-report' },
    }]);
  });

  it('reports a seek position without consuming the watched-time heartbeat', async () => {
    let now = 0;
    let snapshot = { positionSeconds: 0, durationSeconds: 100 };
    const reports: PlaybackReport[] = [];
    const controller = new PlaybackLifecycleController({
      now: () => now,
      read: () => snapshot,
      report: async (report) => { reports.push(report); },
      reportIDFactory: () => 'watched-report',
    });

    controller.playing();
    now = 5_000;
    snapshot = { positionSeconds: 42, durationSeconds: 100 };
    await controller.seeked();
    now = 15_000;
    await controller.periodicFlush();

    expect(reports).toEqual([
      { positionSeconds: 42, progressPercent: 42 },
      {
        positionSeconds: 42,
        progressPercent: 42,
        heartbeat: { watched_seconds_delta: 15, report_id: 'watched-report' },
      },
    ]);
  });

  it('still persists the completed position when pause flushes immediately before ended', async () => {
    let now = 0;
    let snapshot = { positionSeconds: 0, durationSeconds: 60 };
    const reports: PlaybackReport[] = [];
    const controller = new PlaybackLifecycleController({
      now: () => now,
      read: () => snapshot,
      report: async (report) => { reports.push(report); },
      reportIDFactory: () => 'pause-report',
    });

    controller.playing();
    now = 15_000;
    snapshot = { positionSeconds: 15, durationSeconds: 60 };
    await controller.pause();
    snapshot = { positionSeconds: 60, durationSeconds: 60 };
    await controller.ended();

    expect(reports).toEqual([
      {
        positionSeconds: 15,
        progressPercent: 25,
        heartbeat: { watched_seconds_delta: 15, report_id: 'pause-report' },
      },
      { positionSeconds: 60, progressPercent: 100 },
    ]);
  });

  it('retries the identical full request after a heartbeat report fails', async () => {
    let now = 0;
    let snapshot = { positionSeconds: 0, durationSeconds: 60 };
    const attempts: PlaybackReport[] = [];
    const controller = new PlaybackLifecycleController({
      now: () => now,
      read: () => snapshot,
      report: async (report) => {
        attempts.push(report);
        if (attempts.length === 1) throw new Error('offline');
      },
      reportIDFactory: () => 'stable-report',
    });

    controller.playing();
    now = 15_000;
    snapshot = { positionSeconds: 15, durationSeconds: 60 };
    await controller.periodicFlush();
    now = 30_000;
    snapshot = { positionSeconds: 30, durationSeconds: 60 };
    await controller.periodicFlush();

    expect(attempts).toHaveLength(2);
    expect(attempts[1]).toEqual(attempts[0]);
    expect(attempts[1]).toEqual({
      positionSeconds: 15,
      progressPercent: 25,
      heartbeat: { watched_seconds_delta: 15, report_id: 'stable-report' },
    });
  });

  it('attempts both the active request and newly sampled seconds during pagehide', async () => {
    let now = 0;
    let releaseRequest!: () => void;
    const requestPending = new Promise<void>((resolve) => { releaseRequest = resolve; });
    const terminalReports: PlaybackReport[] = [];
    let nextID = 0;
    const controller = new PlaybackLifecycleController({
      now: () => now,
      read: () => ({ positionSeconds: now / 1_000, durationSeconds: 100 }),
      report: async () => requestPending,
      terminalReport: (report) => { terminalReports.push(report); },
      reportIDFactory: () => `report-${++nextID}`,
    });

    controller.playing();
    now = 15_000;
    const periodic = controller.periodicFlush();
    await Promise.resolve();
    now = 20_000;
    await controller.pagehide();

    expect(terminalReports).toEqual([
      {
        positionSeconds: 15,
        progressPercent: 15,
        heartbeat: { watched_seconds_delta: 15, report_id: 'report-1' },
      },
      {
        positionSeconds: 20,
        progressPercent: 20,
        heartbeat: { watched_seconds_delta: 5, report_id: 'report-2' },
      },
    ]);

    releaseRequest();
    await periodic;
  });

  it('resumes lifecycle accounting after a bfcache pageshow', async () => {
    let now = 0;
    const reports: PlaybackReport[] = [];
    let nextID = 0;
    const controller = new PlaybackLifecycleController({
      now: () => now,
      read: () => ({ positionSeconds: now / 1_000, durationSeconds: 100 }),
      report: async (report) => { reports.push(report); },
      reportIDFactory: () => `report-${++nextID}`,
    });

    controller.playing();
    now = 5_000;
    await controller.pagehide();
    await controller.pageshow();
    now = 20_000;
    controller.playing();
    now = 23_000;
    await controller.pause();

    expect(reports.map((report) => report.heartbeat?.watched_seconds_delta)).toEqual([5, 3]);
  });

  it('retries a failed terminal payload after bfcache restore with the same report ID', async () => {
    let now = 0;
    const terminalAttempts: PlaybackReport[] = [];
    const retryAttempts: PlaybackReport[] = [];
    const controller = new PlaybackLifecycleController({
      now: () => now,
      read: () => ({ positionSeconds: 5, durationSeconds: 100 }),
      report: async (report) => { retryAttempts.push(report); },
      terminalReport: async (report) => {
        terminalAttempts.push(report);
        throw new Error('keepalive failed');
      },
      reportIDFactory: () => 'terminal-stable-id',
    });

    controller.playing();
    now = 5_000;
    await controller.pagehide();
    await Promise.resolve();
    await controller.pageshow();

    expect(terminalAttempts).toEqual([{
      positionSeconds: 5,
      progressPercent: 5,
      heartbeat: { watched_seconds_delta: 5, report_id: 'terminal-stable-id' },
    }]);
    expect(retryAttempts).toEqual(terminalAttempts);
  });
});
