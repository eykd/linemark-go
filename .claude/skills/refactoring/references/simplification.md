# Simplification Refactorings

## Inline Function

**When**: Function body is as clear as name, or just delegates.

**Steps**:

1. Check not polymorphic (not overridden)
2. Find all callers
3. Replace each call with body
4. Remove function
5. Test after each replacement

```typescript
// Before: Unnecessary indirection
function rating(driver: Driver): number {
  return moreThanFiveLateDeliveries(driver) ? 2 : 1;
}

function moreThanFiveLateDeliveries(driver: Driver): boolean {
  return driver.numberOfLateDeliveries > 5;
}

// After
function rating(driver: Driver): number {
  return driver.numberOfLateDeliveries > 5 ? 2 : 1;
}
```

## Inline Class

**When**: Class no longer justifies its existence after refactoring.

**Steps**:

1. Move all methods and fields to target class
2. Update references to use target
3. Remove empty class

```typescript
// Before: TrackingInformation has become trivial
class Shipment {
  get trackingInfo(): string {
    return this._trackingInformation.display;
  }
  get trackingInformation(): TrackingInformation {
    return this._trackingInformation;
  }
}

class TrackingInformation {
  get display(): string {
    return `${this.shippingCompany}: ${this.trackingNumber}`;
  }
}

// After: Merged into Shipment
class Shipment {
  private shippingCompany: string;
  private trackingNumber: string;

  get trackingInfo(): string {
    return `${this.shippingCompany}: ${this.trackingNumber}`;
  }
}
```

## Remove Dead Code

**When**: Code is never executed. Trust version control for history.

**Steps**:

1. Use tools/IDE to find unused code
2. Delete it
3. Test

```typescript
// Before: hasDiscount is never called
class Order {
  get total(): number {
    return this.items.reduce((s, i) => s + i.price, 0);
  }

  // Dead code - delete it
  hasDiscount(): boolean {
    return this.discountCode !== null;
  }
}

// After
class Order {
  get total(): number {
    return this.items.reduce((s, i) => s + i.price, 0);
  }
}
```

## Collapse Hierarchy

**When**: Subclass and superclass are too similar.

**Steps**:

1. Choose which class to remove
2. Pull up or push down features to merge
3. Remove empty class
4. Update references

```typescript
// Before: Employee and Salesman have almost identical behavior
class Employee {
  protected name: string;
  protected salary: number;
  // ... many shared methods
}

class Salesman extends Employee {
  get bonus(): number {
    return this.sales * 0.1;
  }
}

// After: If all employees can have sales, collapse
class Employee {
  private name: string;
  private salary: number;
  private sales: number = 0;

  get bonus(): number {
    return this.sales * 0.1;
  }
}
```

## Remove Middle Man

**When**: Class mostly just delegates to another class.

**Steps**:

1. Expose delegate object
2. For each delegating method, adjust client to call delegate directly
3. Remove delegating methods

```typescript
// Before: Person just delegates to Department
class Person {
  get department(): Department {
    return this._department;
  }
  get manager(): Employee {
    return this._department.manager;
  }
  get budget(): number {
    return this._department.budget;
  }
  get headCount(): number {
    return this._department.headCount;
  }
}

// After: Let clients talk to Department directly
class Person {
  get department(): Department {
    return this._department;
  }
}

// Client
const manager = person.department.manager;
```

## Replace Superclass/Subclass with Delegate

**When**: Inheritance isn't true is-a relationship; composition is better.

```typescript
// Before: Stack inherits List but isn't really a List
class Stack extends List {
  push(item: T): void {
    this.append(item);
  }
  pop(): T {
    return this.removeLast();
  }
}

// After: Stack uses List via delegation
class Stack<T> {
  private storage = new List<T>();

  push(item: T): void {
    this.storage.append(item);
  }
  pop(): T {
    return this.storage.removeLast();
  }
  // Note: List methods like get(), set() are NOT exposed
}
```

## Substitute Algorithm

**When**: Simpler algorithm becomes apparent.

**Steps**:

1. Ensure algorithm is in separate function
2. Prepare new algorithm
3. Test new algorithm independently
4. Replace old with new
5. Test

```typescript
// Before: Complex search
function findPerson(people: Person[]): Person | undefined {
  for (const p of people) {
    if (p.name === 'Don') return p;
    if (p.name === 'John') return p;
    if (p.name === 'Kent') return p;
  }
  return undefined;
}

// After: Simpler approach
function findPerson(people: Person[]): Person | undefined {
  const candidates = ['Don', 'John', 'Kent'];
  return people.find((p) => candidates.includes(p.name));
}
```

## Signs of Speculative Generality

Remove these if not actually used:

- Abstract classes with only one subclass
- Unused parameters
- Methods/classes only called by tests
- Type parameters with only one instantiation
- Interfaces with single implementation (and no planned extensions)
