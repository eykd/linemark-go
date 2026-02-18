# UTC Storage Patterns

Store all timestamps as UTC ISO 8601 strings for portability and consistency.

## ISO 8601 UTC Format

```
2025-01-15T16:00:00.000Z
│         │         │
│         │         └─ Z = Zulu time (UTC)
│         └─ Time: 16:00:00.000
└─ Date: 2025-01-15
```

**Why ISO 8601 UTC:**

- Unambiguous - no timezone confusion
- Sortable - string comparison works correctly
- Portable - works across all databases, languages, and systems
- Parseable - `new Date(isoString)` works everywhere

## Entity Timestamps

```typescript
interface NextAction {
  id: string;
  text: string;
  resurfaceAt: string | null; // UTC ISO: "2025-01-15T16:00:00.000Z"
  createdAt: string; // UTC ISO: "2025-01-15T08:30:00.000Z"
  updatedAt: string; // UTC ISO: "2025-01-15T12:45:00.000Z"
  doneAt: string | null; // UTC ISO or null
}

interface Session {
  id: string;
  startedAt: string; // UTC ISO
  expiresAt: string; // UTC ISO
  closedAt: string | null;
}
```

## Creating Timestamps

### Current Time

```typescript
// Simple - get current UTC ISO string
const now = new Date().toISOString();
// "2025-01-15T16:30:45.123Z"

// In a function that accepts injected time
function createAction(text: string, currentTime: Date): NextAction {
  return {
    id: crypto.randomUUID(),
    text,
    resurfaceAt: null,
    createdAt: currentTime.toISOString(),
    updatedAt: currentTime.toISOString(),
    doneAt: null,
  };
}
```

### Future Timestamps

```typescript
// Add hours
function addHours(isoString: string, hours: number): string {
  const date = new Date(isoString);
  date.setTime(date.getTime() + hours * 60 * 60 * 1000);
  return date.toISOString();
}

// Add days
function addDays(isoString: string, days: number): string {
  const date = new Date(isoString);
  date.setUTCDate(date.getUTCDate() + days);
  return date.toISOString();
}

// Example: Session expires in 1 hour
const session = {
  startedAt: now,
  expiresAt: addHours(now, 1),
};
```

### Specific Time in Display Timezone

When you need a specific time in a display timezone (e.g., "8am PT tomorrow"):

```typescript
function createTimestampAt(
  baseUtc: Date,
  hour: number,
  daysFromNow: number,
  timezone: string
): string {
  // Calculate the target date in the display timezone
  const targetDate = new Date(baseUtc);
  targetDate.setUTCDate(targetDate.getUTCDate() + daysFromNow);

  // Get date components in target timezone
  const components = getComponentsInTimezone(targetDate, timezone);

  // Create UTC time for that hour in target timezone
  return createUtcFromTimezoneComponents(
    components.year,
    components.month,
    components.day,
    hour,
    timezone
  ).toISOString();
}

// "Tomorrow at 8am PT" in UTC
const tomorrowMorning = createTimestampAt(new Date(), 8, 1, 'America/Los_Angeles');
```

## Database Storage

### D1/SQLite

Store as TEXT columns:

```sql
CREATE TABLE next_actions (
  id TEXT PRIMARY KEY,
  text TEXT NOT NULL,
  resurface_at TEXT,           -- UTC ISO string
  created_at TEXT NOT NULL,    -- UTC ISO string
  updated_at TEXT NOT NULL,    -- UTC ISO string
  done_at TEXT                 -- UTC ISO string or NULL
);

-- Insert
INSERT INTO next_actions (id, text, resurface_at, created_at, updated_at)
VALUES ('act-123', 'Review PR', '2025-01-16T16:00:00.000Z', '2025-01-15T08:00:00.000Z', '2025-01-15T08:00:00.000Z');

-- Query by time range (string comparison works!)
SELECT * FROM next_actions
WHERE resurface_at >= '2025-01-15T00:00:00.000Z'
  AND resurface_at < '2025-01-16T00:00:00.000Z';

-- Order by time (string sort works!)
SELECT * FROM next_actions
WHERE resurface_at IS NOT NULL
ORDER BY resurface_at ASC;
```

### Repository Pattern

```typescript
interface NextActionRepository {
  save(action: NextAction): Promise<void>;
  findById(id: string): Promise<NextAction | null>;
  findResurfacingBefore(utcIso: string): Promise<NextAction[]>;
}

class D1NextActionRepository implements NextActionRepository {
  constructor(private db: D1Database) {}

  async findResurfacingBefore(utcIso: string): Promise<NextAction[]> {
    // Direct string comparison - no timezone conversion needed
    const result = await this.db
      .prepare('SELECT * FROM next_actions WHERE resurface_at <= ? AND done_at IS NULL')
      .bind(utcIso)
      .all();

    return result.results.map(this.mapToEntity);
  }

  private mapToEntity(row: Record<string, unknown>): NextAction {
    return {
      id: row.id as string,
      text: row.text as string,
      resurfaceAt: row.resurface_at as string | null,
      createdAt: row.created_at as string,
      updatedAt: row.updated_at as string,
      doneAt: row.done_at as string | null,
    };
  }
}
```

## Comparing Timestamps

### String Comparison

ISO 8601 UTC strings sort lexicographically:

```typescript
const earlier = '2025-01-15T08:00:00.000Z';
const later = '2025-01-15T16:00:00.000Z';

// String comparison works!
console.log(earlier < later); // true
console.log(earlier > later); // false

// Sorting works!
const times = [later, earlier];
times.sort(); // ['2025-01-15T08:00:00.000Z', '2025-01-15T16:00:00.000Z']
```

### Millisecond Comparison

For complex comparisons, convert to milliseconds:

```typescript
function isExpired(expiresAt: string, currentTime: string): boolean {
  return new Date(expiresAt).getTime() <= new Date(currentTime).getTime();
}

function minutesBetween(startIso: string, endIso: string): number {
  const startMs = new Date(startIso).getTime();
  const endMs = new Date(endIso).getTime();
  return Math.floor((endMs - startMs) / (60 * 1000));
}
```

### Same Day Check

Check if two UTC timestamps fall on the same day in a specific timezone:

```typescript
function isSameDayInTimezone(iso1: string, iso2: string, timezone: string): boolean {
  const components1 = getComponentsInTimezone(new Date(iso1), timezone);
  const components2 = getComponentsInTimezone(new Date(iso2), timezone);

  return (
    components1.year === components2.year &&
    components1.month === components2.month &&
    components1.day === components2.day
  );
}

// Check if action resurfaces "today" in PT
const isToday = isSameDayInTimezone(
  action.resurfaceAt,
  new Date().toISOString(),
  'America/Los_Angeles'
);
```

## Null Handling

Use `null` for optional timestamps, not empty strings:

```typescript
interface NextAction {
  resurfaceAt: string | null; // null = no scheduled resurface
  doneAt: string | null; // null = not completed
}

// Creating incomplete action
const action: NextAction = {
  id: 'act-123',
  text: 'Review PR',
  resurfaceAt: null, // Not scheduled yet
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  doneAt: null, // Not done
};

// Completing action
function markComplete(action: NextAction, currentTime: Date): NextAction {
  return {
    ...action,
    doneAt: currentTime.toISOString(),
    updatedAt: currentTime.toISOString(),
  };
}
```

## Validation

Validate ISO 8601 UTC format:

```typescript
const UTC_ISO_REGEX = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z$/;

function isValidUtcIso(value: string): boolean {
  if (!UTC_ISO_REGEX.test(value)) {
    return false;
  }

  const date = new Date(value);
  return !isNaN(date.getTime());
}

// Usage
if (!isValidUtcIso(input.resurfaceAt)) {
  throw new Error('resurfaceAt must be a valid UTC ISO 8601 string');
}
```

## Anti-Patterns

### Don't store timezone offsets

```typescript
// ❌ BAD - timezone offset can become stale (DST changes)
const stored = '2025-01-15T08:00:00-08:00';

// ✅ GOOD - always UTC
const stored = '2025-01-15T16:00:00.000Z';
```

### Don't store local time strings

```typescript
// ❌ BAD - ambiguous, not portable
const stored = '2025-01-15 08:00:00';

// ✅ GOOD - unambiguous UTC
const stored = '2025-01-15T16:00:00.000Z';
```

### Don't store Unix timestamps as primary format

```typescript
// ❌ BAD - less readable, harder to debug
const stored = 1736956800000;

// ✅ GOOD - human readable, still sortable
const stored = '2025-01-15T16:00:00.000Z';
```

## Summary

- Store all timestamps as ISO 8601 UTC strings ending in `Z`
- Use `new Date().toISOString()` to get current time
- Use UTC methods (`setUTCDate`, `setUTCHours`) for date math
- String comparison works for sorting and range queries
- Use `null` for optional timestamps
- Validate input with regex + Date parsing
