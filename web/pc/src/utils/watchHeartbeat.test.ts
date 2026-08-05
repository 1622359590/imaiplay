import { describe, expect, it } from 'vitest';
import { WatchHeartbeat } from './watchHeartbeat';

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
