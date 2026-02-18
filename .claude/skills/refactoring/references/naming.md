# Naming Refactorings

## Change Function Declaration (Rename Function)

**When**: Function name doesn't clearly communicate purpose.

**Steps** (Simple):

1. Change function name in declaration
2. Update all call sites
3. Test

**Steps** (Migration for published APIs):

1. Create new function with better name
2. Have old function delegate to new
3. Migrate callers gradually
4. Remove old function

```typescript
// Before
function circum(radius: number): number {
  return 2 * Math.PI * radius;
}

// After
function circumference(radius: number): number {
  return 2 * Math.PI * radius;
}
```

## Rename Variable

**When**: Variable name is unclear or misleading.

**Steps**:

1. If widely used, consider Encapsulate Variable first
2. Change name in declaration
3. Update all references
4. Use IDE automated renaming when available

```typescript
// Before
const a = height * width;

// After
const area = height * width;
```

## Rename Field

**When**: Field name doesn't match current understanding of data.

**Steps**:

1. If record has limited scope, rename directly
2. Otherwise, use Encapsulate Record first
3. Rename private field
4. Adjust accessors

```typescript
// Before
interface Organization {
  name: string;
  ctry: string;
}

// After
interface Organization {
  name: string;
  country: string;
}
```

## Comments → Better Names

**When**: Comments describe what code does (not why).

Comments are often a sign of unclear code. Instead of documenting unclear code, make it clear through better naming.

```typescript
// Before
// Check if customer is eligible for discount
if (customer.age > 65 || customer.membershipYears > 10) {
  // Apply senior or loyalty discount
  total = total * 0.9;
}

// After
const isEligibleForDiscount = customer.age > 65 || customer.membershipYears > 10;
if (isEligibleForDiscount) {
  total = applyDiscount(total, LOYALTY_DISCOUNT_RATE);
}

// OR even better: Extract Function
if (isEligibleForDiscount(customer)) {
  total = applyDiscount(total, LOYALTY_DISCOUNT_RATE);
}
```

## Good Naming Principles

1. **Use domain language** — Match terms stakeholders use
2. **Reveal intent** — Name after what, not how
3. **Be specific** — `customerCount` not `data`
4. **Avoid abbreviations** — `circumference` not `circum`
5. **Keep comments for "why"** — Code should explain "what"

| Bad Name    | Good Name        | Reason                   |
| ----------- | ---------------- | ------------------------ |
| `d`         | `elapsedDays`    | Reveals what it measures |
| `list`      | `customers`      | Reveals domain meaning   |
| `flag`      | `isActive`       | Reveals purpose          |
| `temp`      | `subtotal`       | Reveals business concept |
| `doStuff()` | `calculateTax()` | Reveals operation        |
| `data`      | `userProfile`    | Reveals content type     |
