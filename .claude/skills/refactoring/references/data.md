# Data Refactorings

## Split Variable

**When**: Variable is assigned multiple times for different purposes (except loop variables and accumulators).

**Steps**:

1. Change variable name at first declaration
2. Change references up to next assignment
3. Declare new variable at next assignment
4. Repeat for each assignment

```typescript
// Before: temp used for two different things
let temp = 2 * (height + width);
console.log(temp); // perimeter
temp = height * width;
console.log(temp); // area

// After: Separate variables for separate purposes
const perimeter = 2 * (height + width);
console.log(perimeter);
const area = height * width;
console.log(area);
```

## Replace Derived Variable with Query

**When**: Variable can be calculated from other data; eliminates mutable state.

**Steps**:

1. Identify all update points
2. Create function to calculate value
3. Use assertion to verify calculation matches
4. Replace variable reads with function call
5. Remove variable and updates

```typescript
// Before: discountedTotal updated manually
class ProductionPlan {
  private _production: number = 0;
  private _discountedTotal: number = 0;

  addAdjustment(amount: number): void {
    this._production += amount;
    this._discountedTotal += amount * 0.9;
  }

  get discountedTotal(): number {
    return this._discountedTotal;
  }
}

// After: Calculate when needed
class ProductionPlan {
  private _production: number = 0;

  addAdjustment(amount: number): void {
    this._production += amount;
  }

  get discountedTotal(): number {
    return this._production * 0.9;
  }
}
```

## Change Reference to Value

**When**: Object should have value semantics (compared by content, not identity).

**Steps**:

1. Check object is/can be immutable
2. Create equals method based on fields
3. Remove setter methods
4. Consider making fields readonly

```typescript
// Before: TelephoneNumber compared by reference
class Person {
  private _telephoneNumber: TelephoneNumber;

  set officeAreaCode(value: string) {
    this._telephoneNumber.areaCode = value;
  }
}

// After: Immutable value object
class TelephoneNumber {
  constructor(
    readonly areaCode: string,
    readonly number: string
  ) {}

  equals(other: TelephoneNumber): boolean {
    return this.areaCode === other.areaCode && this.number === other.number;
  }
}

class Person {
  private _telephoneNumber: TelephoneNumber;

  set officeAreaCode(value: string) {
    this._telephoneNumber = new TelephoneNumber(value, this._telephoneNumber.number);
  }
}
```

## Change Value to Reference

**When**: Need to share single instance so updates are seen everywhere.

**Steps**:

1. Create repository for instances
2. Ensure constructor can look up correct instance
3. Change factory to return reference from repository

```typescript
// Before: Each Order has its own Customer copy
class Order {
  constructor(data: OrderData) {
    this._customer = new Customer(data.customerId);
  }
}

// After: Orders share Customer references
const customerRepository = new Map<string, Customer>();

function findCustomer(id: string): Customer {
  if (!customerRepository.has(id)) {
    customerRepository.set(id, new Customer(id));
  }
  return customerRepository.get(id)!;
}

class Order {
  constructor(data: OrderData) {
    this._customer = findCustomer(data.customerId);
  }
}
```

## Replace Loop with Pipeline

**When**: Loop processes collection; pipeline operations are clearer.

**Steps**:

1. Create variable for loop result
2. Convert each loop operation to pipeline operation
3. Remove loop

```typescript
// Before
const names: string[] = [];
for (const person of people) {
  if (person.job === 'programmer') {
    names.push(person.name);
  }
}

// After
const names = people.filter((p) => p.job === 'programmer').map((p) => p.name);
```

## Common Pipeline Operations

| Loop Pattern               | Pipeline Operation              |
| -------------------------- | ------------------------------- |
| Filter items               | `.filter(predicate)`            |
| Transform items            | `.map(transform)`               |
| Find single item           | `.find(predicate)`              |
| Check if any match         | `.some(predicate)`              |
| Check if all match         | `.every(predicate)`             |
| Accumulate to single value | `.reduce(accumulator, initial)` |
| Flatten nested arrays      | `.flatMap(transform)`           |
