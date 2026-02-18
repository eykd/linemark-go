# Encapsulation Refactorings

## Encapsulate Variable

**When**: Data is accessed widely; need to control access, add validation, or prepare for restructuring.

**Steps**:

1. Create getter and setter functions
2. Replace all direct references with function calls
3. Consider making data private
4. Test

```typescript
// Before
let defaultOwner = { firstName: 'Martin', lastName: 'Fowler' };
spaceship.owner = defaultOwner;

// After
let _defaultOwner = { firstName: 'Martin', lastName: 'Fowler' };

function defaultOwner(): Owner {
  return _defaultOwner;
}

function setDefaultOwner(owner: Owner): void {
  _defaultOwner = owner;
}

spaceship.owner = defaultOwner();
```

## Encapsulate Record

**When**: Data structures (objects, hashes) need controlled access.

**Steps**:

1. Create class with private field for record
2. Provide methods to get/set values
3. Replace record usage with class
4. Consider making immutable

```typescript
// Before
const organization = { name: 'Acme', country: 'US' };
organization.name = 'New Name';

// After
class Organization {
  constructor(private data: { name: string; country: string }) {}

  get name(): string {
    return this.data.name;
  }
  set name(value: string) {
    this.data.name = value;
  }
  get country(): string {
    return this.data.country;
  }
}
```

## Encapsulate Collection

**When**: Collection is exposed directly, allowing uncontrolled modification.

**Steps**:

1. Add methods to add/remove items
2. Return read-only view or copy from getter
3. Never return mutable reference

```typescript
// Before
class Person {
  courses: Course[] = [];
}
person.courses.push(newCourse);

// After
class Person {
  private _courses: Course[] = [];

  get courses(): readonly Course[] {
    return this._courses;
  }

  addCourse(course: Course): void {
    this._courses.push(course);
  }

  removeCourse(course: Course): void {
    const index = this._courses.indexOf(course);
    if (index > -1) this._courses.splice(index, 1);
  }
}
```

## Replace Primitive with Object

**When**: Primitive types (string, number) represent domain concepts.

**Steps**:

1. Create class for the value
2. Replace primitive with object
3. Move related behavior into class
4. Consider making immutable

```typescript
// Before
function deliveryDate(order: Order): Date {
  if (order.priority === 'high') {
    return addDays(order.placedOn, 1);
  }
  return addDays(order.placedOn, 3);
}

// After
class Priority {
  constructor(private readonly value: string) {
    if (!['low', 'normal', 'high', 'rush'].includes(value)) {
      throw new Error(`Invalid priority: ${value}`);
    }
  }

  higherThan(other: Priority): boolean {
    const levels = ['low', 'normal', 'high', 'rush'];
    return levels.indexOf(this.value) > levels.indexOf(other.value);
  }
}
```

## Hide Delegate

**When**: Client navigates through one object to get to another (message chains).

**Steps**:

1. Create delegating method on server for each delegate method needed
2. Adjust client to call server
3. Remove client's knowledge of delegate

```typescript
// Before
const manager = person.department.manager;

// After (in Person class)
get manager(): Employee {
  return this.department.manager;
}

// Client code
const manager = person.manager;
```

## Introduce Special Case (Null Object)

**When**: Same null/special case checks appear everywhere.

**Steps**:

1. Create special-case class implementing same interface
2. Return special-case instance instead of null
3. Move special-case behavior into special class

```typescript
// Before
function customerName(site: Site): string {
  const customer = site.customer;
  if (customer === null) return 'occupant';
  return customer.name;
}

// After
class UnknownCustomer implements Customer {
  get name(): string {
    return 'occupant';
  }
  get billingPlan(): BillingPlan {
    return BillingPlan.basic();
  }
}

function customerName(site: Site): string {
  return site.customer.name; // Works for real and unknown customers
}
```
