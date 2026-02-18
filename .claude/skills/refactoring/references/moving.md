# Moving Refactorings

## Move Function

**When**: Function uses elements from another context more than its own (Feature Envy).

**Steps**:

1. Examine function's context usage
2. Copy function to target context
3. Adjust for new location (rename, change parameters)
4. Set up delegation from old to new
5. Test, then remove old or leave as delegation

```typescript
// Before: trackSummary uses totalDistance more than its own class
class GPS {
  trackSummary(points: Point[]): Summary {
    const totalDistance = this.calculateDistance(points);
    const pace = totalDistance / points.length;
    return { distance: totalDistance, pace };
  }

  private calculateDistance(points: Point[]): number {
    return points.reduce((sum, p, i) => (i === 0 ? 0 : sum + p.distanceTo(points[i - 1])), 0);
  }
}

// After: Move to where data lives
class Track {
  constructor(private points: Point[]) {}

  get totalDistance(): number {
    return this.points.reduce(
      (sum, p, i) => (i === 0 ? 0 : sum + p.distanceTo(this.points[i - 1])),
      0
    );
  }

  get pace(): number {
    return this.totalDistance / this.points.length;
  }
}
```

## Move Field

**When**: Field is used more by another class, or data structures are too coupled.

**Steps**:

1. If public, use Encapsulate Variable first
2. Create field in target
3. Adjust references to use target field
4. Remove source field

```typescript
// Before: discount relates more to CustomerContract
class Customer {
  discountRate: number;
  contract: CustomerContract;
}

// After
class Customer {
  contract: CustomerContract;

  get discountRate(): number {
    return this.contract.discountRate;
  }
}

class CustomerContract {
  discountRate: number;
}
```

## Move Statements into Function

**When**: Same code appears in multiple callers of a function.

**Steps**:

1. If statements aren't adjacent, use Slide Statements first
2. Copy statements into function body
3. Test
4. Remove statements from callers

```typescript
// Before
function renderPerson(person: Person): string {
  const result: string[] = [];
  result.push(`<p>${person.name}</p>`);
  result.push(renderPhoto(person.photo));
  result.push(`<p>title: ${person.photo.title}</p>`);
  return result.join('\n');
}

// After: title rendering moved into renderPhoto
function renderPhoto(photo: Photo): string {
  return [`<img src="${photo.url}">`, `<p>title: ${photo.title}</p>`].join('\n');
}
```

## Move Statements to Callers

**When**: Function does too much; some behavior should vary by caller.

**Steps**:

1. Use Slide Statements to move varying code to function exit
2. Copy varying code to each caller
3. Remove from function

```typescript
// Before: emitPhotoData always outputs location, but not all callers want it
function emitPhotoData(photo: Photo): string {
  return `<p>title: ${photo.title}</p>\n<p>location: ${photo.location}</p>`;
}

// After
function emitPhotoData(photo: Photo): string {
  return `<p>title: ${photo.title}</p>`;
}

// Callers that need location add it themselves
console.log(emitPhotoData(photo) + `\n<p>location: ${photo.location}</p>`);
```

## Slide Statements

**When**: Related code is scattered; group for Extract Function or clarity.

**Steps**:

1. Identify target position
2. Check for dependencies that would break if moved
3. Move code to target
4. Test

```typescript
// Before
const pricingPlan = retrievePricingPlan();
const order = retrieveOrder();
let charge: number;
const chargePerUnit = pricingPlan.unit;

// After: Group pricing-related code
const pricingPlan = retrievePricingPlan();
const chargePerUnit = pricingPlan.unit;
const order = retrieveOrder();
let charge: number;
```

## Replace Inline Code with Function Call

**When**: Code duplicates logic that exists in a library or elsewhere.

```typescript
// Before
let hasDiscount = false;
for (const customer of customers) {
  if (customer.isPremium) {
    hasDiscount = true;
    break;
  }
}

// After
const hasDiscount = customers.some((c) => c.isPremium);
```

## Split Loop

**When**: Loop does multiple unrelated things.

**Steps**:

1. Copy loop
2. Remove different operations from each copy
3. Test
4. Consider Extract Function on each loop

```typescript
// Before
let youngest = Infinity;
let totalSalary = 0;
for (const p of people) {
  if (p.age < youngest) youngest = p.age;
  totalSalary += p.salary;
}

// After
let youngest = Infinity;
for (const p of people) {
  if (p.age < youngest) youngest = p.age;
}

let totalSalary = 0;
for (const p of people) {
  totalSalary += p.salary;
}

// Even better: Use collection methods
const youngest = Math.min(...people.map((p) => p.age));
const totalSalary = people.reduce((sum, p) => sum + p.salary, 0);
```
