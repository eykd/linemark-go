# Inheritance Refactorings

## Pull Up Method

**When**: Methods in subclasses do the same thing.

**Steps**:

1. Check methods are identical (or make them so)
2. Check method signatures match
3. Create method in superclass
4. Remove from subclasses
5. Test

```typescript
// Before: Duplicate in subclasses
class Employee {
  protected annualCost(): number {
    /* abstract */
  }
}

class Engineer extends Employee {
  get annualCost(): number {
    return this.monthlyCost * 12;
  }
}

class Salesman extends Employee {
  get annualCost(): number {
    return this.monthlyCost * 12;
  }
}

// After: Pull up to superclass
abstract class Employee {
  abstract get monthlyCost(): number;

  get annualCost(): number {
    return this.monthlyCost * 12;
  }
}
```

## Pull Up Field

**When**: Subclasses have same field.

**Steps**:

1. Check fields are used similarly
2. Create field in superclass
3. Remove from subclasses

```typescript
// Before
class Engineer extends Employee {
  protected name: string;
}

class Salesman extends Employee {
  protected name: string;
}

// After
abstract class Employee {
  protected name: string;
}

class Engineer extends Employee {}
class Salesman extends Employee {}
```

## Pull Up Constructor Body

**When**: Subclass constructors have common code.

**Steps**:

1. Create superclass constructor if needed
2. Move common statements to superclass
3. Call super() from subclasses

```typescript
// Before: Repeated initialization
class Employee {
  protected name: string;
}

class Engineer extends Employee {
  constructor(
    name: string,
    private specialization: string
  ) {
    super();
    this.name = name;
  }
}

class Manager extends Employee {
  constructor(
    name: string,
    private department: string
  ) {
    super();
    this.name = name;
  }
}

// After
class Employee {
  constructor(protected name: string) {}
}

class Engineer extends Employee {
  constructor(
    name: string,
    private specialization: string
  ) {
    super(name);
  }
}

class Manager extends Employee {
  constructor(
    name: string,
    private department: string
  ) {
    super(name);
  }
}
```

## Push Down Method

**When**: Method only relevant to specific subclass.

**Steps**:

1. Copy method to subclass(es) that need it
2. Remove from superclass

```typescript
// Before: Only Salesman uses quota
class Employee {
  get quota(): number {
    return this._quota;
  }
}

// After
class Employee {}

class Salesman extends Employee {
  get quota(): number {
    return this._quota;
  }
}
```

## Push Down Field

**When**: Field only used by specific subclass.

```typescript
// Before
class Employee {
  protected quota: number; // Only used by Salesman
}

// After
class Employee {}

class Salesman extends Employee {
  private quota: number;
}
```

## Replace Subclass with Delegate

**When**: Inheritance doesn't fit; need more flexibility.

**Why**:

- Can only inherit once
- Inheritance couples tightly
- Subclass relationship isn't true is-a

**Steps**:

1. Create delegate class for subclass behavior
2. Add delegate field to superclass
3. Move subclass methods to delegate
4. Replace subclass with factory that configures delegate

```typescript
// Before: Booking inheritance limits flexibility
class Booking {
  constructor(
    protected show: Show,
    protected date: Date
  ) {}

  get hasTalkback(): boolean {
    return false;
  }
  get basePrice(): number {
    return this.show.price;
  }
}

class PremiumBooking extends Booking {
  constructor(
    show: Show,
    date: Date,
    private extras: Extras
  ) {
    super(show, date);
  }

  get hasTalkback(): boolean {
    return this.show.hasOwnProperty('talkback') && !this.isPeakDay;
  }

  get basePrice(): number {
    return Math.round(super.basePrice + this.extras.premiumFee);
  }
}

// After: Delegate provides flexibility
class Booking {
  private premiumDelegate?: PremiumBookingDelegate;

  constructor(
    protected show: Show,
    protected date: Date
  ) {}

  bePremium(extras: Extras): void {
    this.premiumDelegate = new PremiumBookingDelegate(this, extras);
  }

  get hasTalkback(): boolean {
    return this.premiumDelegate?.hasTalkback ?? false;
  }

  get basePrice(): number {
    const base = this.show.price;
    return this.premiumDelegate?.adjustPrice(base) ?? base;
  }
}

class PremiumBookingDelegate {
  constructor(
    private host: Booking,
    private extras: Extras
  ) {}

  get hasTalkback(): boolean {
    return this.host.show.hasOwnProperty('talkback') && !this.host.isPeakDay;
  }

  adjustPrice(base: number): number {
    return Math.round(base + this.extras.premiumFee);
  }
}
```

## Extract Superclass

**When**: Classes share features; hierarchy is appropriate.

**Steps**:

1. Create empty superclass
2. Make classes extend it
3. Pull up common features one by one

```typescript
// Before: Department and Employee share features
class Employee {
  constructor(
    protected name: string,
    protected annualCost: number
  ) {}
  get monthlySpend(): number {
    return this.annualCost / 12;
  }
}

class Department {
  constructor(
    protected name: string,
    protected staff: Employee[]
  ) {}
  get totalAnnualCost(): number {
    return this.staff.reduce((sum, e) => sum + e.annualCost, 0);
  }
  get monthlySpend(): number {
    return this.totalAnnualCost / 12;
  }
}

// After
abstract class Party {
  constructor(protected name: string) {}
  abstract get annualCost(): number;
  get monthlySpend(): number {
    return this.annualCost / 12;
  }
}

class Employee extends Party {
  constructor(
    name: string,
    private _annualCost: number
  ) {
    super(name);
  }
  get annualCost(): number {
    return this._annualCost;
  }
}

class Department extends Party {
  constructor(
    name: string,
    private staff: Employee[]
  ) {
    super(name);
  }
  get annualCost(): number {
    return this.staff.reduce((sum, e) => sum + e.annualCost, 0);
  }
}
```

## When to Prefer Delegation Over Inheritance

| Use Inheritance When               | Use Delegation When                 |
| ---------------------------------- | ----------------------------------- |
| True is-a relationship             | Has-a or uses-a relationship        |
| Subclass uses most of superclass   | Only needs part of interface        |
| Subclass is truly a specialization | Need runtime flexibility            |
| Hierarchy is stable                | Behavior might change independently |
| Only need single inheritance       | Need multiple "parents"             |
