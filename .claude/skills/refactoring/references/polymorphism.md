# Polymorphism Refactorings

## Replace Conditional with Polymorphism

**When**: Same switch/if-else checks type in multiple places.

**Steps**:

1. Create class hierarchy if needed
2. Use factory for object creation
3. Move conditional logic into polymorphic method
4. Replace conditional with method call

```typescript
// Before: Switch on type in multiple places
function plumage(bird: Bird): string {
  switch (bird.type) {
    case 'EuropeanSwallow':
      return 'average';
    case 'AfricanSwallow':
      return bird.numberOfCoconuts > 2 ? 'tired' : 'average';
    case 'NorwegianBlueParrot':
      return bird.voltage > 100 ? 'scorched' : 'beautiful';
    default:
      return 'unknown';
  }
}

function airSpeed(bird: Bird): number {
  switch (bird.type) {
    case 'EuropeanSwallow':
      return 35;
    case 'AfricanSwallow':
      return 40 - 2 * bird.numberOfCoconuts;
    case 'NorwegianBlueParrot':
      return bird.isNailed ? 0 : 10 + bird.voltage / 10;
    default:
      return 0;
  }
}

// After: Polymorphic classes
abstract class Bird {
  abstract get plumage(): string;
  abstract get airSpeed(): number;
}

class EuropeanSwallow extends Bird {
  get plumage(): string {
    return 'average';
  }
  get airSpeed(): number {
    return 35;
  }
}

class AfricanSwallow extends Bird {
  constructor(private numberOfCoconuts: number) {
    super();
  }
  get plumage(): string {
    return this.numberOfCoconuts > 2 ? 'tired' : 'average';
  }
  get airSpeed(): number {
    return 40 - 2 * this.numberOfCoconuts;
  }
}

class NorwegianBlueParrot extends Bird {
  constructor(
    private voltage: number,
    private isNailed: boolean
  ) {
    super();
  }
  get plumage(): string {
    return this.voltage > 100 ? 'scorched' : 'beautiful';
  }
  get airSpeed(): number {
    return this.isNailed ? 0 : 10 + this.voltage / 10;
  }
}

// Factory
function createBird(data: BirdData): Bird {
  switch (data.type) {
    case 'EuropeanSwallow':
      return new EuropeanSwallow();
    case 'AfricanSwallow':
      return new AfricanSwallow(data.numberOfCoconuts);
    case 'NorwegianBlueParrot':
      return new NorwegianBlueParrot(data.voltage, data.isNailed);
    default:
      throw new Error(`Unknown bird type: ${data.type}`);
  }
}
```

## Decompose Conditional

**When**: Complex condition obscures intent; first step toward polymorphism.

**Steps**:

1. Extract condition into function with clear name
2. Extract then-branch into function
3. Extract else-branch into function

```typescript
// Before
if (date < SUMMER_START || date > SUMMER_END) {
  charge = quantity * winterRate + winterServiceCharge;
} else {
  charge = quantity * summerRate;
}

// After
if (isSummer(date)) {
  charge = summerCharge(quantity);
} else {
  charge = winterCharge(quantity);
}

function isSummer(date: Date): boolean {
  return date >= SUMMER_START && date <= SUMMER_END;
}

function summerCharge(quantity: number): number {
  return quantity * summerRate;
}

function winterCharge(quantity: number): number {
  return quantity * winterRate + winterServiceCharge;
}
```

## Replace Type Code with Subclasses

**When**: Type code affects behavior; enables polymorphism.

**Steps**:

1. Self-encapsulate type code if not already
2. Create subclass for each type code value
3. Create factory to return appropriate subclass
4. Replace type code checks with polymorphic methods

```typescript
// Before: Type code as field
class Employee {
  constructor(private type: 'engineer' | 'salesman' | 'manager') {}

  get bonus(): number {
    switch (this.type) {
      case 'engineer':
        return this.salary * 0.1;
      case 'salesman':
        return this.sales * 0.15;
      case 'manager':
        return this.salary * 0.2 + this.teamBonus;
    }
  }
}

// After: Type hierarchy
abstract class Employee {
  abstract get bonus(): number;
}

class Engineer extends Employee {
  get bonus(): number {
    return this.salary * 0.1;
  }
}

class Salesman extends Employee {
  get bonus(): number {
    return this.sales * 0.15;
  }
}

class Manager extends Employee {
  get bonus(): number {
    return this.salary * 0.2 + this.teamBonus;
  }
}

function createEmployee(type: string, data: EmployeeData): Employee {
  switch (type) {
    case 'engineer':
      return new Engineer(data);
    case 'salesman':
      return new Salesman(data);
    case 'manager':
      return new Manager(data);
    default:
      throw new Error(`Unknown type: ${type}`);
  }
}
```

## Introduce Assertion

**When**: Assumptions about state should be explicit.

**Steps**:

1. Identify assumption that must be true
2. Add assertion checking assumption
3. Use for invariants, not user input validation

```typescript
// Before: Implicit assumption about discount
function applyDiscount(price: number, discountRate: number): number {
  return price - price * discountRate;
}

// After: Explicit assertion
function applyDiscount(price: number, discountRate: number): number {
  console.assert(
    discountRate >= 0 && discountRate <= 1,
    `Discount rate must be 0-1, got ${discountRate}`
  );
  return price - price * discountRate;
}
```

## When NOT to Use Polymorphism

- **One-off conditional**: If switch appears only once, leave it
- **Simple cases**: Don't create hierarchy for 2-3 simple cases
- **External data**: Type info comes from JSON/DB, keep factory switch
- **Performance critical**: Virtual dispatch has overhead
