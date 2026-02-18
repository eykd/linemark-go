# Value Objects

Value objects are immutable, defined by their attributes, and have no identity.

## Table of Contents

- [Characteristics](#characteristics)
- [Patterns](#patterns)
- [Common Value Objects](#common-value-objects)
- [Equality and Comparison](#equality-and-comparison)
- [Testing Value Objects](#testing-value-objects)

## Characteristics

| Characteristic   | Description                                        |
| ---------------- | -------------------------------------------------- |
| Immutable        | State never changes after creation                 |
| No identity      | Two instances with same values are interchangeable |
| Self-validating  | Constructor rejects invalid values                 |
| Side-effect free | Methods return new instances                       |

## Patterns

### Basic Value Object

```typescript
export class Email {
  private readonly _value: string;

  private constructor(value: string) {
    this._value = value;
  }

  static create(value: string): Email {
    const normalized = value.toLowerCase().trim();
    if (!Email.isValid(normalized)) {
      throw new Error('Invalid email format');
    }
    return new Email(normalized);
  }

  private static isValid(email: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  }

  get value(): string {
    return this._value;
  }

  get domain(): string {
    return this._value.split('@')[1];
  }

  equals(other: Email): boolean {
    return this._value === other._value;
  }
}
```

### Enum-Style Value Object

```typescript
export class TaskStatus {
  private static readonly PENDING = new TaskStatus('pending');
  private static readonly IN_PROGRESS = new TaskStatus('in_progress');
  private static readonly COMPLETED = new TaskStatus('completed');

  private constructor(private readonly _value: string) {}

  static pending(): TaskStatus {
    return TaskStatus.PENDING;
  }
  static inProgress(): TaskStatus {
    return TaskStatus.IN_PROGRESS;
  }
  static completed(): TaskStatus {
    return TaskStatus.COMPLETED;
  }

  static fromString(value: string): TaskStatus {
    switch (value) {
      case 'pending':
        return TaskStatus.PENDING;
      case 'in_progress':
        return TaskStatus.IN_PROGRESS;
      case 'completed':
        return TaskStatus.COMPLETED;
      default:
        throw new Error(`Invalid status: ${value}`);
    }
  }

  get value(): string {
    return this._value;
  }
  get isPending(): boolean {
    return this === TaskStatus.PENDING;
  }
  get isCompleted(): boolean {
    return this === TaskStatus.COMPLETED;
  }

  canTransitionTo(target: TaskStatus): boolean {
    if (this === TaskStatus.PENDING) {
      return target === TaskStatus.IN_PROGRESS;
    }
    if (this === TaskStatus.IN_PROGRESS) {
      return target === TaskStatus.COMPLETED || target === TaskStatus.PENDING;
    }
    return false; // Completed is terminal
  }
}
```

### Composite Value Object

```typescript
export class Money {
  private constructor(
    public readonly amount: number,
    public readonly currency: string
  ) {}

  static of(amount: number, currency: string): Money {
    if (!Number.isFinite(amount)) {
      throw new Error('Amount must be a finite number');
    }
    if (amount < 0) {
      throw new Error('Amount cannot be negative');
    }
    if (!/^[A-Z]{3}$/.test(currency)) {
      throw new Error('Currency must be 3-letter ISO code');
    }
    // Round to 2 decimal places
    return new Money(Math.round(amount * 100) / 100, currency);
  }

  static zero(currency: string): Money {
    return Money.of(0, currency);
  }

  add(other: Money): Money {
    this.assertSameCurrency(other);
    return Money.of(this.amount + other.amount, this.currency);
  }

  subtract(other: Money): Money {
    this.assertSameCurrency(other);
    if (other.amount > this.amount) {
      throw new Error('Cannot subtract: would result in negative amount');
    }
    return Money.of(this.amount - other.amount, this.currency);
  }

  multiply(factor: number): Money {
    if (factor < 0) throw new Error('Factor cannot be negative');
    return Money.of(this.amount * factor, this.currency);
  }

  private assertSameCurrency(other: Money): void {
    if (this.currency !== other.currency) {
      throw new Error(`Currency mismatch: ${this.currency} vs ${other.currency}`);
    }
  }

  equals(other: Money): boolean {
    return this.amount === other.amount && this.currency === other.currency;
  }

  toString(): string {
    return `${this.currency} ${this.amount.toFixed(2)}`;
  }
}
```

## Common Value Objects

### ID Value Object

```typescript
export class TaskId {
  private constructor(public readonly value: string) {}

  static create(): TaskId {
    return new TaskId(crypto.randomUUID());
  }

  static from(value: string): TaskId {
    if (!value || value.trim().length === 0) {
      throw new Error('TaskId cannot be empty');
    }
    return new TaskId(value);
  }

  equals(other: TaskId): boolean {
    return this.value === other.value;
  }
}
```

### Date Range Value Object

```typescript
export class DateRange {
  private constructor(
    public readonly start: Date,
    public readonly end: Date
  ) {}

  static create(start: Date, end: Date): DateRange {
    if (end < start) {
      throw new Error('End date must be after start date');
    }
    return new DateRange(start, end);
  }

  contains(date: Date): boolean {
    return date >= this.start && date <= this.end;
  }

  overlaps(other: DateRange): boolean {
    return this.start <= other.end && this.end >= other.start;
  }

  get durationInDays(): number {
    return Math.ceil((this.end.getTime() - this.start.getTime()) / (1000 * 60 * 60 * 24));
  }
}
```

### Address Value Object

```typescript
export class Address {
  private constructor(
    public readonly street: string,
    public readonly city: string,
    public readonly postalCode: string,
    public readonly country: string
  ) {}

  static create(props: {
    street: string;
    city: string;
    postalCode: string;
    country: string;
  }): Address {
    if (!props.street?.trim()) throw new Error('Street is required');
    if (!props.city?.trim()) throw new Error('City is required');
    if (!props.postalCode?.trim()) throw new Error('Postal code is required');
    if (!props.country?.trim()) throw new Error('Country is required');

    return new Address(
      props.street.trim(),
      props.city.trim(),
      props.postalCode.trim().toUpperCase(),
      props.country.trim()
    );
  }

  equals(other: Address): boolean {
    return (
      this.street === other.street &&
      this.city === other.city &&
      this.postalCode === other.postalCode &&
      this.country === other.country
    );
  }

  format(): string {
    return `${this.street}, ${this.city}, ${this.postalCode}, ${this.country}`;
  }
}
```

## Equality and Comparison

Always implement `equals()`:

```typescript
// By value comparison
equals(other: Money): boolean {
  return this.amount === other.amount && this.currency === other.currency;
}

// In collections
const prices = [Money.of(10, "USD"), Money.of(20, "USD")];
const target = Money.of(10, "USD");
const found = prices.find(p => p.equals(target)); // Works
```

## Testing Value Objects

```typescript
describe('Money', () => {
  describe('creation', () => {
    it('creates money with valid amount and currency', () => {
      const money = Money.of(100, 'USD');
      expect(money.amount).toBe(100);
      expect(money.currency).toBe('USD');
    });

    it('rounds to 2 decimal places', () => {
      const money = Money.of(10.999, 'USD');
      expect(money.amount).toBe(11);
    });

    it('rejects negative amounts', () => {
      expect(() => Money.of(-10, 'USD')).toThrow('cannot be negative');
    });

    it('rejects invalid currency codes', () => {
      expect(() => Money.of(10, 'US')).toThrow('3-letter ISO');
    });
  });

  describe('operations', () => {
    it('adds same currency', () => {
      const a = Money.of(10, 'USD');
      const b = Money.of(20, 'USD');
      expect(a.add(b).amount).toBe(30);
    });

    it('rejects adding different currencies', () => {
      const usd = Money.of(10, 'USD');
      const eur = Money.of(10, 'EUR');
      expect(() => usd.add(eur)).toThrow('Currency mismatch');
    });
  });

  describe('equality', () => {
    it('equals same value', () => {
      const a = Money.of(10, 'USD');
      const b = Money.of(10, 'USD');
      expect(a.equals(b)).toBe(true);
    });

    it('not equals different amount', () => {
      const a = Money.of(10, 'USD');
      const b = Money.of(20, 'USD');
      expect(a.equals(b)).toBe(false);
    });
  });
});
```
