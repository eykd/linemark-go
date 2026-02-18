---
name: portable-datetime
description: 'Write portable datetime handling code that stores in UTC, calculates in UTC, and displays in a configurable timezone. Use when: (1) Working with dates/times that must work across timezones, (2) Storing timestamps in databases, (3) Displaying times to users in their locale, (4) Scheduling future events, (5) Writing timezone-independent tests, or (6) Converting between UTC and local time.'
---

# Portable Datetime Handling

Handle datetimes correctly across timezones using the UTC Sandwich pattern: convert to UTC at input boundaries, work exclusively in UTC within your domain, and convert to display timezones only at output boundaries.

## The UTC Sandwich Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                    System Boundary                          │
│  ┌─────────────┐                          ┌─────────────┐  │
│  │   INPUT     │    ┌──────────────┐      │   OUTPUT    │  │
│  │  (any tz)   │───▶│  DOMAIN      │─────▶│  (display   │  │
│  │             │    │  (UTC only)  │      │   tz)       │  │
│  │  Convert    │    │              │      │  Convert    │  │
│  │  to UTC     │    │  Store UTC   │      │  from UTC   │  │
│  └─────────────┘    │  Calc UTC    │      └─────────────┘  │
│                     │  Compare UTC │                        │
│                     └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

**Why this matters:** Tests run identically anywhere, database queries work without timezone math, comparisons are simple, DST transitions don't corrupt logic, and scheduling is unambiguous.

## Quick Reference

| What you need                      | Where to look                                                                     |
| ---------------------------------- | --------------------------------------------------------------------------------- |
| Store timestamps in database       | [utc-storage.md](references/utc-storage.md) - ISO 8601 patterns                   |
| Display times to users             | [timezone-conversion.md](references/timezone-conversion.md) - Intl.DateTimeFormat |
| Add/subtract time, schedule events | [common-operations.md](references/common-operations.md) - Calculation recipes     |
| Write deterministic tests          | [testing-time.md](references/testing-time.md) - Time injection patterns           |

## Decision Framework

```typescript
// STORING? → ISO 8601 UTC strings
resurfaceAt: string; // "2025-01-15T16:00:00.000Z"

// CALCULATING? → UTC methods
tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);

// DISPLAYING? → Convert at boundary only
formatForTimezone(utcString, 'America/Los_Angeles');
```

## Essential Patterns

**Store as UTC ISO strings** - See [utc-storage.md](references/utc-storage.md)

```typescript
startedAt: now.toISOString(); // "2025-01-15T16:00:00.000Z"
```

**Calculate in UTC** - See [common-operations.md](references/common-operations.md)

```typescript
tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);
tomorrow.setUTCHours(8, 0, 0, 0);
```

**Display at boundaries** - See [timezone-conversion.md](references/timezone-conversion.md)

```typescript
new Intl.DateTimeFormat('en-US', {
  timeZone: timezone,
  month: 'short',
  day: 'numeric',
}).format(new Date(utcIsoString));
```

**Inject time for testing** - See [testing-time.md](references/testing-time.md)

```typescript
function computeResurfaceAt(option: DeferOption, currentTimeUtc: Date): string {
  const tomorrow = new Date(currentTimeUtc);
  tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);
  return tomorrow.toISOString();
}

it('computes tomorrow correctly', () => {
  const now = new Date('2025-01-15T17:00:00.000Z');
  expect(computeResurfaceAt({ type: 'TOMORROW' }, now)).toBe('2025-01-16T16:00:00.000Z');
});
```

## Critical Anti-Patterns

```typescript
// ❌ BAD - Local methods depend on server timezone
tomorrow.setDate(tomorrow.getDate() + 1);
// ✅ GOOD - Explicit UTC
tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);

// ❌ BAD - Domain entity knows about display timezone
class Session {
  getDisplayTime(): string {
    return formatInPT(this.startedAt);
  }
}
// ✅ GOOD - Return UTC, convert at boundary
class Session {
  getStartedAt(): string {
    return this.startedAt;
  }
}

// ❌ BAD - Ambiguous timezone parsing
new Date('2025-01-15 08:00:00');
// ✅ GOOD - Explicit UTC
new Date('2025-01-15T08:00:00.000Z');
```

## Summary Rules

1. Store UTC: All persisted timestamps are ISO 8601 UTC strings
2. Calculate UTC: Use `setUTC*` and `getUTC*` methods for date math
3. Convert at boundaries: Only use timezone conversion in adapters/presenters
4. Inject time: Pass `currentTimeUtc: Date` as parameter, never `new Date()` in domain
5. Test with fixed UTC: Use known UTC timestamps, assert UTC outputs

## Reference Documentation

- [utc-storage.md](references/utc-storage.md) - Database schemas, type definitions, parsing, validation
- [timezone-conversion.md](references/timezone-conversion.md) - Display formatting, component extraction, DST handling
- [testing-time.md](references/testing-time.md) - Test fixtures, time injection, mocking, edge cases
- [common-operations.md](references/common-operations.md) - Duration math, scheduling, business days, comparisons

## Related Skills

- `/prefactoring` - Design time-handling abstractions and boundaries
- `/typescript-unit-testing` - Write deterministic datetime tests
