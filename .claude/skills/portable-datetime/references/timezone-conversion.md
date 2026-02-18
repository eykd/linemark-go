# Timezone Conversion Patterns

Convert between UTC and display timezones at system boundaries only.

## The Intl.DateTimeFormat Approach

Use `Intl.DateTimeFormat` for all timezone conversions - it handles DST automatically:

```typescript
const TIMEZONE = 'America/Los_Angeles';

// Format for display
function formatForDisplay(utcIso: string): string {
  const date = new Date(utcIso);

  return new Intl.DateTimeFormat('en-US', {
    timeZone: TIMEZONE,
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  }).format(date);
}

// "Jan 15, 8:00 AM"
```

## Getting Components in a Timezone

Extract year, month, day, hour from a UTC date in a specific timezone:

```typescript
function getComponentsInTimezone(
  date: Date,
  timezone: string
): { year: number; month: number; day: number; hour: number; minute: number; dayOfWeek: number } {
  const formatter = new Intl.DateTimeFormat('en-US', {
    timeZone: timezone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
    weekday: 'short',
  });

  const parts = formatter.formatToParts(date);
  const get = (type: string): string => parts.find((p) => p.type === type)?.value ?? '';

  const weekdays = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

  return {
    year: parseInt(get('year'), 10),
    month: parseInt(get('month'), 10),
    day: parseInt(get('day'), 10),
    hour: parseInt(get('hour'), 10),
    minute: parseInt(get('minute'), 10),
    dayOfWeek: weekdays.indexOf(get('weekday')),
  };
}

// Usage
const utc = new Date('2025-01-15T16:00:00.000Z');
const pt = getComponentsInTimezone(utc, 'America/Los_Angeles');
// { year: 2025, month: 1, day: 15, hour: 8, minute: 0, dayOfWeek: 3 }
// (Wednesday at 8am PT)
```

## Creating UTC from Timezone Components

Convert a time in a specific timezone to UTC:

```typescript
function createUtcFromTimezoneComponents(
  year: number,
  month: number,
  day: number,
  hour: number,
  minute: number,
  timezone: string
): Date {
  // Build a date string
  const dateStr = [year, String(month).padStart(2, '0'), String(day).padStart(2, '0')].join('-');

  const timeStr = [String(hour).padStart(2, '0'), String(minute).padStart(2, '0'), '00'].join(':');

  // Parse as UTC first to get a reference
  const tempUtc = new Date(`${dateStr}T${timeStr}Z`);

  // See what hour this shows in the target timezone
  const tzFormatter = new Intl.DateTimeFormat('en-US', {
    timeZone: timezone,
    hour: '2-digit',
    hour12: false,
  });
  const tzHour = parseInt(tzFormatter.format(tempUtc), 10);

  // Calculate the offset
  let offsetHours = tzHour - hour;

  // Handle day boundary wraparound
  if (offsetHours > 12) offsetHours -= 24;
  if (offsetHours < -12) offsetHours += 24;

  // Adjust to get the correct UTC time
  const result = new Date(tempUtc);
  result.setUTCHours(result.getUTCHours() - offsetHours);
  result.setUTCMinutes(
    result.getUTCMinutes() -
      (minute -
        parseInt(
          new Intl.DateTimeFormat('en-US', {
            timeZone: timezone,
            minute: '2-digit',
          }).format(tempUtc),
          10
        ))
  );

  return result;
}

// "8am PT on Jan 15, 2025" as UTC
const utc = createUtcFromTimezoneComponents(2025, 1, 15, 8, 0, 'America/Los_Angeles');
// 2025-01-15T16:00:00.000Z (PST = UTC-8)
```

## Display Formatting Recipes

### Relative Time Display

```typescript
function formatRelative(utcIso: string, nowUtc: string, timezone: string): string {
  const targetComponents = getComponentsInTimezone(new Date(utcIso), timezone);
  const nowComponents = getComponentsInTimezone(new Date(nowUtc), timezone);

  // Same day?
  if (
    targetComponents.year === nowComponents.year &&
    targetComponents.month === nowComponents.month &&
    targetComponents.day === nowComponents.day
  ) {
    return 'Today';
  }

  // Tomorrow?
  const tomorrow = new Date(nowUtc);
  tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);
  const tomorrowComponents = getComponentsInTimezone(tomorrow, timezone);

  if (
    targetComponents.year === tomorrowComponents.year &&
    targetComponents.month === tomorrowComponents.month &&
    targetComponents.day === tomorrowComponents.day
  ) {
    return 'Tomorrow';
  }

  // Otherwise show date
  const months = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'May',
    'Jun',
    'Jul',
    'Aug',
    'Sep',
    'Oct',
    'Nov',
    'Dec',
  ];
  return `${months[targetComponents.month - 1]} ${targetComponents.day}`;
}
```

### Time with AM/PM

```typescript
function formatTimeAmPm(utcIso: string, timezone: string): string {
  const components = getComponentsInTimezone(new Date(utcIso), timezone);

  const hour12 = components.hour % 12 || 12;
  const ampm = components.hour >= 12 ? 'pm' : 'am';
  const minute = components.minute > 0 ? `:${String(components.minute).padStart(2, '0')}` : '';

  return `${hour12}${minute}${ampm}`;
}

// "8am", "1:30pm", "12pm"
```

### Full Date and Time

```typescript
function formatFullDateTime(utcIso: string, timezone: string): string {
  return new Intl.DateTimeFormat('en-US', {
    timeZone: timezone,
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  }).format(new Date(utcIso));
}

// "Wed, Jan 15, 8:00 AM"
```

### Day of Week

```typescript
function getDayOfWeekName(utcIso: string, timezone: string): string {
  return new Intl.DateTimeFormat('en-US', {
    timeZone: timezone,
    weekday: 'long',
  }).format(new Date(utcIso));
}

// "Wednesday"
```

## DST Handling

`Intl.DateTimeFormat` handles DST automatically. Be aware of edge cases:

### DST Transitions

```typescript
// Winter (PST = UTC-8)
const winter = '2025-01-15T16:00:00.000Z'; // 8am PT
const winterPt = getComponentsInTimezone(new Date(winter), 'America/Los_Angeles');
// hour: 8

// Summer (PDT = UTC-7)
const summer = '2025-07-15T15:00:00.000Z'; // 8am PT
const summerPt = getComponentsInTimezone(new Date(summer), 'America/Los_Angeles');
// hour: 8

// Note: Same local hour, different UTC hours!
```

### Creating Times Near DST Boundary

When scheduling events, DST transitions can cause issues:

```typescript
// Spring forward: 2am PT doesn't exist on March 9, 2025
// Fall back: 1am PT happens twice on November 2, 2025

// Safe approach: Use Intl to handle the conversion
function scheduleAt(targetHour: number, targetDate: Date, timezone: string): string {
  const components = getComponentsInTimezone(targetDate, timezone);

  return createUtcFromTimezoneComponents(
    components.year,
    components.month,
    components.day,
    targetHour,
    0,
    timezone
  ).toISOString();
}
```

## Presentation Layer Integration

### Slack Adapter Example

```typescript
// src/adapters/slack-message-formatter.ts
const USER_TIMEZONE = 'America/Los_Angeles';

export function formatActionCardMeta(action: NextAction): string {
  if (!action.resurfaceAt) {
    return '';
  }

  const relative = formatRelative(action.resurfaceAt, new Date().toISOString(), USER_TIMEZONE);

  const time = formatTimeAmPm(action.resurfaceAt, USER_TIMEZONE);

  return `_Resurfaces: ${relative} at ${time}_`;
}
```

### HTTP Response Example

```typescript
// src/adapters/http-response-formatter.ts
export function formatSessionResponse(
  session: ReviewSession,
  userTimezone: string
): SessionResponse {
  return {
    id: session.id,
    startedAt: session.startedAt, // Keep UTC for API consumers
    startedAtDisplay: formatFullDateTime(session.startedAt, userTimezone),
    expiresAt: session.expiresAt,
    expiresAtDisplay: formatFullDateTime(session.expiresAt, userTimezone),
  };
}
```

## Common Timezone Identifiers

```typescript
// US Timezones
'America/Los_Angeles'; // Pacific (PT)
'America/Denver'; // Mountain (MT)
'America/Chicago'; // Central (CT)
'America/New_York'; // Eastern (ET)

// Other common
'Europe/London'; // UK (GMT/BST)
'Europe/Paris'; // Central Europe (CET/CEST)
'Asia/Tokyo'; // Japan (JST)
'Australia/Sydney'; // Australia Eastern (AEST/AEDT)

// UTC
'UTC'; // Coordinated Universal Time
```

## Anti-Patterns

### Don't hardcode offsets

```typescript
// ❌ BAD - hardcoded offset ignores DST
function ptToUtc(ptHour: number): number {
  return ptHour + 8; // Only correct in winter!
}

// ✅ GOOD - use Intl for correct offset
const utc = createUtcFromTimezoneComponents(2025, 1, 15, 8, 0, 'America/Los_Angeles');
```

### Don't use Date.prototype timezone methods

```typescript
// ❌ BAD - uses system timezone
const hour = date.getHours();
const day = date.getDay();

// ✅ GOOD - explicit timezone
const components = getComponentsInTimezone(date, 'America/Los_Angeles');
const hour = components.hour;
const day = components.dayOfWeek;
```

### Don't convert in domain logic

```typescript
// ❌ BAD - timezone logic in domain
class NextAction {
  getDisplayTime(): string {
    return formatForDisplay(this.resurfaceAt); // Wrong place!
  }
}

// ✅ GOOD - convert in adapter only
// Domain returns UTC
const utcTime = action.resurfaceAt;
// Adapter converts for display
const displayTime = formatForDisplay(utcTime, userTimezone);
```

## Summary

- Use `Intl.DateTimeFormat` for all timezone conversions
- Extract components with `formatToParts()` for calculations
- Create UTC from timezone components using offset calculation
- Convert only at system boundaries (adapters/presenters)
- DST is handled automatically by Intl
- Never hardcode timezone offsets
