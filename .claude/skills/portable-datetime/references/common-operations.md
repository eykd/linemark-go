# Common Datetime Operations

Recipes for everyday datetime tasks using UTC storage and portable timezone handling.

## Adding Time Durations

### Add Hours

```typescript
function addHours(utcIso: string, hours: number): string {
  const date = new Date(utcIso);
  date.setTime(date.getTime() + hours * 60 * 60 * 1000);
  return date.toISOString();
}

// 2 hours from now
const later = addHours('2025-01-15T16:00:00.000Z', 2);
// '2025-01-15T18:00:00.000Z'
```

### Add Days

```typescript
function addDays(utcIso: string, days: number): string {
  const date = new Date(utcIso);
  date.setUTCDate(date.getUTCDate() + days);
  return date.toISOString();
}

// Tomorrow same time
const tomorrow = addDays('2025-01-15T16:00:00.000Z', 1);
// '2025-01-16T16:00:00.000Z'
```

### Add Weeks

```typescript
function addWeeks(utcIso: string, weeks: number): string {
  return addDays(utcIso, weeks * 7);
}

// Next week
const nextWeek = addWeeks('2025-01-15T16:00:00.000Z', 1);
// '2025-01-22T16:00:00.000Z'
```

### Add Minutes

```typescript
function addMinutes(utcIso: string, minutes: number): string {
  const date = new Date(utcIso);
  date.setTime(date.getTime() + minutes * 60 * 1000);
  return date.toISOString();
}

// 30 minutes from now
const later = addMinutes('2025-01-15T16:00:00.000Z', 30);
// '2025-01-15T16:30:00.000Z'
```

## Finding Specific Days

### Next Occurrence of Day of Week

```typescript
function getNextDayOfWeek(
  fromUtc: Date,
  targetDay: number, // 0=Sun, 1=Mon, ..., 6=Sat
  timezone: string
): Date {
  const fromComponents = getComponentsInTimezone(fromUtc, timezone);
  const currentDay = fromComponents.dayOfWeek;

  // Calculate days until target
  let daysUntil = (targetDay - currentDay + 7) % 7;
  if (daysUntil === 0) daysUntil = 7; // Always go to next occurrence

  const targetDate = new Date(fromUtc);
  targetDate.setUTCDate(targetDate.getUTCDate() + daysUntil);

  return targetDate;
}

// Next Saturday from Wednesday Jan 15
const nextSat = getNextDayOfWeek(
  new Date('2025-01-15T17:00:00.000Z'),
  6, // Saturday
  'America/Los_Angeles'
);
// Points to Jan 18
```

### Next Monday

```typescript
function getNextMonday(fromUtc: Date, timezone: string): Date {
  return getNextDayOfWeek(fromUtc, 1, timezone);
}
```

### Next Weekday (Mon-Fri)

```typescript
function getNextWeekday(fromUtc: Date, timezone: string): Date {
  const components = getComponentsInTimezone(fromUtc, timezone);
  const dayOfWeek = components.dayOfWeek;

  let daysToAdd: number;
  if (dayOfWeek === 5) {
    // Friday -> Monday
    daysToAdd = 3;
  } else if (dayOfWeek === 6) {
    // Saturday -> Monday
    daysToAdd = 2;
  } else {
    // Sun-Thu -> next day
    daysToAdd = 1;
  }

  const result = new Date(fromUtc);
  result.setUTCDate(result.getUTCDate() + daysToAdd);
  return result;
}
```

## Setting Specific Times

### Set Time in Display Timezone

```typescript
function setTimeInTimezone(baseUtc: Date, hour: number, minute: number, timezone: string): string {
  const components = getComponentsInTimezone(baseUtc, timezone);

  return createUtcFromTimezoneComponents(
    components.year,
    components.month,
    components.day,
    hour,
    minute,
    timezone
  ).toISOString();
}

// Set to 8am PT
const morning = setTimeInTimezone(
  new Date('2025-01-15T17:00:00.000Z'),
  8,
  0,
  'America/Los_Angeles'
);
// '2025-01-15T16:00:00.000Z'
```

### Tomorrow at Specific Time

```typescript
function tomorrowAt(fromUtc: Date, hour: number, minute: number, timezone: string): string {
  const tomorrow = new Date(fromUtc);
  tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);

  return setTimeInTimezone(tomorrow, hour, minute, timezone);
}

// Tomorrow at 8am PT
const tomorrowMorning = tomorrowAt(
  new Date('2025-01-15T17:00:00.000Z'),
  8,
  0,
  'America/Los_Angeles'
);
// '2025-01-16T16:00:00.000Z'
```

### Start of Day in Timezone

```typescript
function startOfDayInTimezone(utcIso: string, timezone: string): string {
  const components = getComponentsInTimezone(new Date(utcIso), timezone);

  return createUtcFromTimezoneComponents(
    components.year,
    components.month,
    components.day,
    0,
    0,
    timezone
  ).toISOString();
}

// Start of Jan 15 in PT
const startOfDay = startOfDayInTimezone('2025-01-15T17:00:00.000Z', 'America/Los_Angeles');
// '2025-01-15T08:00:00.000Z' (midnight PT = 8am UTC)
```

### End of Day in Timezone

```typescript
function endOfDayInTimezone(utcIso: string, timezone: string): string {
  const components = getComponentsInTimezone(new Date(utcIso), timezone);

  return createUtcFromTimezoneComponents(
    components.year,
    components.month,
    components.day,
    23,
    59,
    timezone
  ).toISOString();
}
```

## Comparing Dates

### Is Before/After

```typescript
function isBefore(utcIso1: string, utcIso2: string): boolean {
  return new Date(utcIso1).getTime() < new Date(utcIso2).getTime();
}

function isAfter(utcIso1: string, utcIso2: string): boolean {
  return new Date(utcIso1).getTime() > new Date(utcIso2).getTime();
}

function isBeforeOrEqual(utcIso1: string, utcIso2: string): boolean {
  return new Date(utcIso1).getTime() <= new Date(utcIso2).getTime();
}
```

### Is Same Day in Timezone

```typescript
function isSameDay(utcIso1: string, utcIso2: string, timezone: string): boolean {
  const c1 = getComponentsInTimezone(new Date(utcIso1), timezone);
  const c2 = getComponentsInTimezone(new Date(utcIso2), timezone);

  return c1.year === c2.year && c1.month === c2.month && c1.day === c2.day;
}

// Both on Jan 15 in PT?
const sameDay = isSameDay(
  '2025-01-15T17:00:00.000Z',
  '2025-01-15T23:00:00.000Z',
  'America/Los_Angeles'
);
// true (9am PT and 3pm PT are same day)
```

### Is Today in Timezone

```typescript
function isToday(utcIso: string, nowUtc: Date, timezone: string): boolean {
  return isSameDay(utcIso, nowUtc.toISOString(), timezone);
}
```

### Is Tomorrow in Timezone

```typescript
function isTomorrow(utcIso: string, nowUtc: Date, timezone: string): boolean {
  const tomorrow = new Date(nowUtc);
  tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);

  return isSameDay(utcIso, tomorrow.toISOString(), timezone);
}
```

### Is Weekend in Timezone

```typescript
function isWeekend(utcIso: string, timezone: string): boolean {
  const components = getComponentsInTimezone(new Date(utcIso), timezone);
  return components.dayOfWeek === 0 || components.dayOfWeek === 6;
}
```

## Time Differences

### Minutes Between

```typescript
function minutesBetween(startUtcIso: string, endUtcIso: string): number {
  const startMs = new Date(startUtcIso).getTime();
  const endMs = new Date(endUtcIso).getTime();
  return Math.floor((endMs - startMs) / (60 * 1000));
}
```

### Hours Between

```typescript
function hoursBetween(startUtcIso: string, endUtcIso: string): number {
  return Math.floor(minutesBetween(startUtcIso, endUtcIso) / 60);
}
```

### Days Between

```typescript
function daysBetween(startUtcIso: string, endUtcIso: string): number {
  return Math.floor(hoursBetween(startUtcIso, endUtcIso) / 24);
}
```

### Human-Readable Duration

```typescript
function formatDuration(startUtcIso: string, endUtcIso: string): string {
  const minutes = minutesBetween(startUtcIso, endUtcIso);

  if (minutes < 60) {
    return `${minutes} min`;
  }

  const hours = Math.floor(minutes / 60);
  if (hours < 24) {
    return `${hours}h ${minutes % 60}m`;
  }

  const days = Math.floor(hours / 24);
  return `${days}d ${hours % 24}h`;
}
```

## Working Hours

### Is Within Working Hours

```typescript
interface WorkingHours {
  startHour: number;
  endHour: number;
  timezone: string;
  workDays: number[]; // 1=Mon, 5=Fri
}

const DEFAULT_WORKING_HOURS: WorkingHours = {
  startHour: 9,
  endHour: 17,
  timezone: 'America/Los_Angeles',
  workDays: [1, 2, 3, 4, 5], // Mon-Fri
};

function isWithinWorkingHours(
  utcIso: string,
  hours: WorkingHours = DEFAULT_WORKING_HOURS
): boolean {
  const components = getComponentsInTimezone(new Date(utcIso), hours.timezone);

  // Check day
  if (!hours.workDays.includes(components.dayOfWeek)) {
    return false;
  }

  // Check hour
  return components.hour >= hours.startHour && components.hour < hours.endHour;
}
```

### Next Working Hour

```typescript
function getNextWorkingHour(fromUtc: Date, hours: WorkingHours = DEFAULT_WORKING_HOURS): string {
  const components = getComponentsInTimezone(fromUtc, hours.timezone);

  // If currently in working hours, return current time
  if (
    hours.workDays.includes(components.dayOfWeek) &&
    components.hour >= hours.startHour &&
    components.hour < hours.endHour
  ) {
    return fromUtc.toISOString();
  }

  // Find next working day
  let targetDate = new Date(fromUtc);
  let daysChecked = 0;

  while (daysChecked < 7) {
    const dayComponents = getComponentsInTimezone(targetDate, hours.timezone);

    if (hours.workDays.includes(dayComponents.dayOfWeek)) {
      // If today but after hours, or future working day
      if (daysChecked > 0 || dayComponents.hour >= hours.endHour) {
        // Move to start of this working day
        if (daysChecked === 0) {
          targetDate.setUTCDate(targetDate.getUTCDate() + 1);
        }
        return setTimeInTimezone(targetDate, hours.startHour, 0, hours.timezone);
      }

      // Today before end, but before start
      if (dayComponents.hour < hours.startHour) {
        return setTimeInTimezone(targetDate, hours.startHour, 0, hours.timezone);
      }
    }

    targetDate.setUTCDate(targetDate.getUTCDate() + 1);
    daysChecked++;
  }

  // Fallback (shouldn't reach)
  return fromUtc.toISOString();
}
```

## Scheduling Helpers

### Defer Options

```typescript
type DeferOption = 'LATER_TODAY' | 'THIS_AFTERNOON' | 'TOMORROW' | 'THIS_WEEKEND' | 'NEXT_WEEK';

function computeDeferTime(option: DeferOption, fromUtc: Date, timezone: string): string {
  const components = getComponentsInTimezone(fromUtc, timezone);

  switch (option) {
    case 'LATER_TODAY': {
      // 2 hours from now, but at least 10am
      const targetHour = Math.max(components.hour + 2, 10);
      return setTimeInTimezone(fromUtc, targetHour, 0, timezone);
    }

    case 'THIS_AFTERNOON': {
      // 1pm today
      return setTimeInTimezone(fromUtc, 13, 0, timezone);
    }

    case 'TOMORROW': {
      // 8am tomorrow
      return tomorrowAt(fromUtc, 8, 0, timezone);
    }

    case 'THIS_WEEKEND': {
      // Saturday 8am
      const saturday = getNextDayOfWeek(fromUtc, 6, timezone);
      return setTimeInTimezone(saturday, 8, 0, timezone);
    }

    case 'NEXT_WEEK': {
      // Monday 8am
      const monday = getNextDayOfWeek(fromUtc, 1, timezone);
      return setTimeInTimezone(monday, 8, 0, timezone);
    }
  }
}
```

## Date Range Queries

### Items Due Today

```typescript
function findDueToday(
  items: Array<{ dueAt: string }>,
  nowUtc: Date,
  timezone: string
): Array<{ dueAt: string }> {
  return items.filter((item) => isToday(item.dueAt, nowUtc, timezone));
}
```

### Items Due This Week

```typescript
function findDueThisWeek(
  items: Array<{ dueAt: string }>,
  nowUtc: Date,
  timezone: string
): Array<{ dueAt: string }> {
  const nowComponents = getComponentsInTimezone(nowUtc, timezone);

  // Calculate start of week (Monday)
  const daysFromMonday = (nowComponents.dayOfWeek + 6) % 7;
  const startOfWeek = new Date(nowUtc);
  startOfWeek.setUTCDate(startOfWeek.getUTCDate() - daysFromMonday);
  const weekStart = startOfDayInTimezone(startOfWeek.toISOString(), timezone);

  // End of week (Sunday night)
  const endOfWeek = new Date(startOfWeek);
  endOfWeek.setUTCDate(endOfWeek.getUTCDate() + 6);
  const weekEnd = endOfDayInTimezone(endOfWeek.toISOString(), timezone);

  return items.filter((item) => isAfter(item.dueAt, weekStart) && isBefore(item.dueAt, weekEnd));
}
```

### Overdue Items

```typescript
function findOverdue(items: Array<{ dueAt: string }>, nowUtc: string): Array<{ dueAt: string }> {
  return items.filter((item) => isBefore(item.dueAt, nowUtc));
}
```

## Complete Utility Module

```typescript
// src/utils/datetime.ts

export const DISPLAY_TIMEZONE = 'America/Los_Angeles';

export function getComponentsInTimezone(
  date: Date,
  timezone: string
): { year: number; month: number; day: number; hour: number; minute: number; dayOfWeek: number } {
  // ... implementation from timezone-conversion.md
}

export function createUtcFromTimezoneComponents(
  year: number,
  month: number,
  day: number,
  hour: number,
  minute: number,
  timezone: string
): Date {
  // ... implementation from timezone-conversion.md
}

export function formatForDisplay(utcIso: string, timezone: string): string {
  // ... implementation
}

export function addHours(utcIso: string, hours: number): string {
  /* ... */
}
export function addDays(utcIso: string, days: number): string {
  /* ... */
}
export function isBefore(a: string, b: string): boolean {
  /* ... */
}
export function isAfter(a: string, b: string): boolean {
  /* ... */
}
export function isSameDay(a: string, b: string, tz: string): boolean {
  /* ... */
}
export function isToday(utcIso: string, now: Date, tz: string): boolean {
  /* ... */
}

// ... and other helpers as needed
```

## Summary

Common patterns for datetime operations:

- **Adding time**: Work in milliseconds or use UTC methods
- **Finding days**: Calculate day-of-week in display timezone
- **Setting times**: Convert timezone components to UTC
- **Comparing**: Use milliseconds or ISO string comparison
- **Durations**: Calculate difference in milliseconds, convert to units
- **Working hours**: Check day-of-week and hour in display timezone
- **Scheduling**: Combine day-finding with time-setting helpers
