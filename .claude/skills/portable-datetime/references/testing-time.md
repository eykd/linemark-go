# Deterministic Time Testing

Write tests that produce identical results regardless of when or where they run.

## Core Principle: Inject Time as Dependency

Never call `new Date()` inside domain logic. Always accept time as a parameter:

```typescript
// ❌ BAD - non-deterministic
function createSession(): Session {
  return {
    id: crypto.randomUUID(),
    startedAt: new Date().toISOString(), // Different every time!
    expiresAt: addHours(new Date(), 1),
  };
}

// ✅ GOOD - deterministic
function createSession(id: string, currentTime: Date): Session {
  return {
    id,
    startedAt: currentTime.toISOString(),
    expiresAt: addHours(currentTime, 1).toISOString(),
  };
}

// In tests
const fixedTime = new Date('2025-01-15T16:00:00.000Z');
const session = createSession('session-1', fixedTime);

expect(session.startedAt).toBe('2025-01-15T16:00:00.000Z');
expect(session.expiresAt).toBe('2025-01-15T17:00:00.000Z');
```

## Fixed Test Timestamps

Create a set of well-documented test timestamps:

```typescript
// test/fixtures/test-times.ts

/**
 * Fixed UTC timestamps for deterministic testing.
 * All times are carefully chosen to represent specific scenarios.
 */
export const TEST_TIMES = {
  // Wednesday Jan 15, 2025 at 5pm UTC = 9am PT (PST, UTC-8)
  WEDNESDAY_9AM_PT: '2025-01-15T17:00:00.000Z',

  // Wednesday Jan 15, 2025 at 3pm UTC = 7am PT
  WEDNESDAY_7AM_PT: '2025-01-15T15:00:00.000Z',

  // Thursday Jan 16, 2025 at 4pm UTC = 8am PT
  THURSDAY_8AM_PT: '2025-01-16T16:00:00.000Z',

  // Saturday Jan 18, 2025 at 4pm UTC = 8am PT
  SATURDAY_8AM_PT: '2025-01-18T16:00:00.000Z',

  // Monday Jan 20, 2025 at 4pm UTC = 8am PT
  MONDAY_8AM_PT: '2025-01-20T16:00:00.000Z',

  // Summer time (PDT, UTC-7)
  // Wednesday Jul 16, 2025 at 3pm UTC = 8am PT
  SUMMER_WEDNESDAY_8AM_PT: '2025-07-16T15:00:00.000Z',
};

// Helper to create Date objects
export function getTestTime(key: keyof typeof TEST_TIMES): Date {
  return new Date(TEST_TIMES[key]);
}
```

## Testing Time-Based Logic

### Testing "Tomorrow at 8am PT"

```typescript
import { TEST_TIMES, getTestTime } from '../fixtures/test-times';

describe('computeResurfaceAt', () => {
  it('computes TOMORROW as 8am PT next day', () => {
    // Given: Wednesday at 9am PT
    const now = getTestTime('WEDNESDAY_9AM_PT');

    // When: deferring to tomorrow
    const result = computeResurfaceAt({ type: 'TOMORROW' }, now);

    // Then: Thursday at 8am PT (4pm UTC)
    expect(result).toBe(TEST_TIMES.THURSDAY_8AM_PT);
  });
});
```

### Testing Day-of-Week Logic

```typescript
describe('THIS_WEEKEND calculation', () => {
  it('computes THIS_WEEKEND as Saturday 8am PT from Wednesday', () => {
    // Wednesday -> Saturday = 3 days
    const now = getTestTime('WEDNESDAY_9AM_PT');

    const result = computeResurfaceAt({ type: 'THIS_WEEKEND' }, now);

    expect(result).toBe(TEST_TIMES.SATURDAY_8AM_PT);
  });

  it('computes THIS_WEEKEND as next Saturday when already Saturday', () => {
    // Saturday -> next Saturday = 7 days
    const now = getTestTime('SATURDAY_8AM_PT');

    const result = computeResurfaceAt({ type: 'THIS_WEEKEND' }, now);

    // Saturday Jan 25, 2025 at 8am PT
    expect(result).toBe('2025-01-25T16:00:00.000Z');
  });
});
```

### Testing DST Transitions

```typescript
describe('DST handling', () => {
  it('computes correct UTC for winter (PST)', () => {
    // Winter: PT = UTC-8
    const winter = new Date('2025-01-15T17:00:00.000Z'); // 9am PT

    const result = computeResurfaceAt({ type: 'LATER_TODAY' }, winter);

    // 9am + 2h = 11am PT = 7pm UTC (UTC-8)
    expect(result).toBe('2025-01-15T19:00:00.000Z');
  });

  it('computes correct UTC for summer (PDT)', () => {
    // Summer: PT = UTC-7
    const summer = new Date('2025-07-15T16:00:00.000Z'); // 9am PT

    const result = computeResurfaceAt({ type: 'LATER_TODAY' }, summer);

    // 9am + 2h = 11am PT = 6pm UTC (UTC-7)
    expect(result).toBe('2025-07-15T18:00:00.000Z');
  });
});
```

## Testing Display Formatting

```typescript
describe('formatForDisplay', () => {
  it('formats UTC as PT display time', () => {
    // 4pm UTC = 8am PT (winter)
    const utc = '2025-01-15T16:00:00.000Z';

    const display = formatForDisplay(utc, 'America/Los_Angeles');

    expect(display).toBe('Jan 15, 8:00 AM');
  });

  it('formats 12pm correctly', () => {
    // 8pm UTC = 12pm PT (noon)
    const utc = '2025-01-15T20:00:00.000Z';

    const display = formatTimeAmPm(utc, 'America/Los_Angeles');

    expect(display).toBe('12pm');
  });

  it('formats afternoon times with pm', () => {
    // 9pm UTC = 1pm PT
    const utc = '2025-01-15T21:00:00.000Z';

    const display = formatTimeAmPm(utc, 'America/Los_Angeles');

    expect(display).toBe('1pm');
  });
});
```

## Testing Relative Time

```typescript
describe('formatRelative', () => {
  const timezone = 'America/Los_Angeles';

  it('returns "Today" for same day in PT', () => {
    // Both are Jan 15 in PT
    const target = '2025-01-15T20:00:00.000Z'; // 12pm PT
    const now = '2025-01-15T17:00:00.000Z'; // 9am PT

    expect(formatRelative(target, now, timezone)).toBe('Today');
  });

  it('returns "Tomorrow" for next day in PT', () => {
    const target = '2025-01-16T16:00:00.000Z'; // Jan 16 8am PT
    const now = '2025-01-15T17:00:00.000Z'; // Jan 15 9am PT

    expect(formatRelative(target, now, timezone)).toBe('Tomorrow');
  });

  it('handles day boundary in PT correctly', () => {
    // UTC: both on Jan 16
    // PT: now is late Jan 15, target is Jan 16
    const now = '2025-01-16T06:00:00.000Z'; // Jan 15, 10pm PT
    const target = '2025-01-16T16:00:00.000Z'; // Jan 16, 8am PT

    expect(formatRelative(target, now, timezone)).toBe('Tomorrow');
  });
});
```

## Using Fake Timers (When Necessary)

For testing code that internally uses `Date.now()` or `new Date()`:

```typescript
import { vi, beforeEach, afterEach } from 'vitest';

describe('session expiry', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2025-01-15T16:00:00.000Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('expires after timeout', () => {
    const session = createSessionWithCurrentTime();

    // Advance time by 1 hour
    vi.advanceTimersByTime(60 * 60 * 1000);

    expect(isExpired(session)).toBe(true);
  });
});
```

**Note:** Prefer injecting time over fake timers when possible. Fake timers are global and can cause unexpected interactions.

## Test Data Builders with Time

```typescript
// test/builders/action-builder.ts
export class ActionBuilder {
  private data = {
    id: 'action-1',
    text: 'Test action',
    resurfaceAt: null as string | null,
    createdAt: TEST_TIMES.WEDNESDAY_9AM_PT,
    updatedAt: TEST_TIMES.WEDNESDAY_9AM_PT,
    doneAt: null as string | null,
  };

  withId(id: string): this {
    this.data.id = id;
    return this;
  }

  withResurfaceAt(utcIso: string): this {
    this.data.resurfaceAt = utcIso;
    return this;
  }

  resurfacingTomorrow(): this {
    this.data.resurfaceAt = TEST_TIMES.THURSDAY_8AM_PT;
    return this;
  }

  resurfacingSaturday(): this {
    this.data.resurfaceAt = TEST_TIMES.SATURDAY_8AM_PT;
    return this;
  }

  completed(): this {
    this.data.doneAt = this.data.updatedAt;
    return this;
  }

  build(): NextAction {
    return { ...this.data };
  }
}

// Usage
const action = new ActionBuilder().withId('action-123').resurfacingTomorrow().build();
```

## Common Test Scenarios

### Time Zone Edge Cases

```typescript
// Test late night PT (next day in UTC)
it('handles late PT that is next day in UTC', () => {
  // 11pm PT on Jan 15 = 7am UTC on Jan 16
  const lateNight = '2025-01-16T07:00:00.000Z';
  const components = getComponentsInTimezone(new Date(lateNight), 'America/Los_Angeles');

  expect(components.day).toBe(15); // Still Jan 15 in PT
  expect(components.hour).toBe(23); // 11pm PT
});

// Test early morning PT (previous day in UTC)
it('handles early PT that spans midnight UTC', () => {
  // 1am PT on Jan 16 = 9am UTC on Jan 16
  const earlyMorning = '2025-01-16T09:00:00.000Z';
  const components = getComponentsInTimezone(new Date(earlyMorning), 'America/Los_Angeles');

  expect(components.day).toBe(16);
  expect(components.hour).toBe(1);
});
```

### Leap Year and Month Boundaries

```typescript
it('handles leap year February correctly', () => {
  // Feb 28, 2024 (leap year) + 1 day = Feb 29
  const feb28 = new Date('2024-02-28T16:00:00.000Z');
  const tomorrow = new Date(feb28);
  tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);

  expect(tomorrow.getUTCDate()).toBe(29);
  expect(tomorrow.getUTCMonth()).toBe(1); // February
});

it('handles month boundary correctly', () => {
  // Jan 31 + 1 day = Feb 1
  const jan31 = new Date('2025-01-31T16:00:00.000Z');
  const nextDay = new Date(jan31);
  nextDay.setUTCDate(nextDay.getUTCDate() + 1);

  expect(nextDay.getUTCDate()).toBe(1);
  expect(nextDay.getUTCMonth()).toBe(1); // February
});
```

## Anti-Patterns

### Don't rely on system timezone in tests

```typescript
// ❌ BAD - depends on test runner's timezone
it('shows correct hour', () => {
  const date = new Date('2025-01-15T16:00:00.000Z');
  expect(date.getHours()).toBe(8); // Only passes in PT!
});

// ✅ GOOD - explicit timezone
it('shows correct hour in PT', () => {
  const date = new Date('2025-01-15T16:00:00.000Z');
  const components = getComponentsInTimezone(date, 'America/Los_Angeles');
  expect(components.hour).toBe(8);
});
```

### Don't use today's date in tests

```typescript
// ❌ BAD - different results on different days
it('formats as today', () => {
  const now = new Date();
  const target = new Date(now);
  target.setHours(now.getHours() + 2);

  expect(formatRelative(target.toISOString(), now.toISOString(), tz)).toBe('Today');
});

// ✅ GOOD - fixed dates
it('formats as today', () => {
  const now = '2025-01-15T17:00:00.000Z';
  const target = '2025-01-15T19:00:00.000Z';

  expect(formatRelative(target, now, 'America/Los_Angeles')).toBe('Today');
});
```

### Don't assume timezone offset

```typescript
// ❌ BAD - assumes PST (UTC-8), fails in PDT
it('converts to PT', () => {
  const utc = '2025-07-15T16:00:00.000Z';
  const ptHour = 16 - 8; // Wrong in summer!
  expect(formatHour(utc)).toBe(ptHour);
});

// ✅ GOOD - use Intl, works year-round
it('converts to PT', () => {
  const utc = '2025-07-15T16:00:00.000Z';
  const components = getComponentsInTimezone(new Date(utc), 'America/Los_Angeles');
  expect(components.hour).toBe(9); // PDT = UTC-7
});
```

## Summary

- Inject time as a parameter, never call `new Date()` in domain logic
- Create well-documented fixed test timestamps
- Test both winter (PST) and summer (PDT) scenarios
- Use explicit timezone in assertions
- Test edge cases: midnight boundaries, DST transitions, leap years
- Prefer time injection over fake timers
- Never rely on the test runner's system timezone
