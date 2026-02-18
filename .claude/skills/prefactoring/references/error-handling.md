# Error Handling Strategies

**Purpose**: Apply prefactoring principles for robust error handling, meaningful messages, and resilient systems.

## When to Use

- Designing error handling strategies
- Implementing catch blocks
- Creating user-facing error messages
- Building resilient external integrations

## Core Principles

### Never Be Silent

Every error must be reported. Never swallow exceptions without handling.

```typescript
// Bad: Silent failure
try {
  await sendNotification(user);
} catch {
  // Error ignored - user never knows
}

// Good: Always report
try {
  await sendNotification(user);
} catch (error) {
  this.logger.error('Notification failed', { userId: user.id, error });
  throw new NotificationFailedError(user.id, error);
}
```

### Report Meaningful User Messages

Error messages describe what users can do, not technical details.

```typescript
// Bad: Technical error exposed
throw new Error('SQLITE_CONSTRAINT: UNIQUE constraint failed');

// Good: User-actionable message
throw new UserFacingError(
  'An account with this email already exists. Please sign in or use a different email.',
  { code: 'EMAIL_EXISTS', technicalDetails: originalError }
);
```

### Consider Failure an Expectation

Design for failure with retries, fallbacks, and graceful degradation.

```typescript
class ResilientNotificationService {
  async notify(user: User, message: Message): Promise<Result<void>> {
    // Retry with backoff
    for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
      const result = await this.tryNotify(user, message);
      if (result.success) return result;
      await this.backoff(attempt);
    }

    // Fallback
    await this.queueForLaterDelivery(user, message);
    return Result.ok();
  }
}
```

## Result Type Pattern

Use a Result type to make errors explicit in the type system.

```typescript
type Result<T, E = Error> = { success: true; value: T } | { success: false; error: E };

function parseEmail(input: string): Result<Email, ValidationError> {
  if (!input.includes('@')) {
    return { success: false, error: new ValidationError('Invalid email') };
  }
  return { success: true, value: new Email(input) };
}

// Usage forces error handling
const result = parseEmail(input);
if (!result.success) {
  return Response.json({ error: result.error.message }, { status: 400 });
}
const email = result.value; // Type-safe access
```

## Decision Matrix

| Situation           | Apply                  | Example                        |
| ------------------- | ---------------------- | ------------------------------ |
| Catch block         | Never Be Silent        | Log and re-throw or handle     |
| User-facing error   | Meaningful Messages    | Explain what user can do       |
| External dependency | Failure as Expectation | Add retry/fallback             |
| Function can fail   | Result Type            | Return Result instead of throw |

## Related References

- [contracts.md](./contracts.md): Validation and interface contracts
- [value-objects.md](./value-objects.md): Domain types with validation
