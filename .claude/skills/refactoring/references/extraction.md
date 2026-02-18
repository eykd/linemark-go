# Extraction Refactorings

## Extract Function

**When**: Code block has clear purpose, appears multiple times, or function is too long.

**Steps**:

1. Create new function named after intent (what, not how)
2. Copy code to new function
3. Identify variables needed—pass as parameters
4. Replace original with function call
5. Test

```typescript
// Before
function printOwing(invoice: Invoice): void {
  let outstanding = 0;
  for (const o of invoice.orders) {
    outstanding += o.amount;
  }

  console.log(`Customer: ${invoice.customer}`);
  console.log(`Amount: ${outstanding}`);
}

// After
function printOwing(invoice: Invoice): void {
  const outstanding = calculateOutstanding(invoice);
  printDetails(invoice, outstanding);
}

function calculateOutstanding(invoice: Invoice): number {
  return invoice.orders.reduce((sum, o) => sum + o.amount, 0);
}

function printDetails(invoice: Invoice, outstanding: number): void {
  console.log(`Customer: ${invoice.customer}`);
  console.log(`Amount: ${outstanding}`);
}
```

## Inline Function

**When**: Function body is as clear as its name, or function is simple delegation.

**Steps**:

1. Check function isn't polymorphic (not overridden)
2. Find all callers
3. Replace each call with function body
4. Remove function definition
5. Test after each replacement

```typescript
// Before
function getRating(driver: Driver): number {
  return moreThanFiveLateDeliveries(driver) ? 2 : 1;
}

function moreThanFiveLateDeliveries(driver: Driver): boolean {
  return driver.lateDeliveries > 5;
}

// After
function getRating(driver: Driver): number {
  return driver.lateDeliveries > 5 ? 2 : 1;
}
```

## Extract Variable

**When**: Complex expression is hard to understand.

**Steps**:

1. Ensure expression has no side effects
2. Declare immutable variable with clear name
3. Assign expression to variable
4. Replace expression with variable reference

```typescript
// Before
return (
  order.quantity * order.itemPrice -
  Math.max(0, order.quantity - 500) * order.itemPrice * 0.05 +
  Math.min(order.quantity * order.itemPrice * 0.1, 100)
);

// After
const basePrice = order.quantity * order.itemPrice;
const quantityDiscount = Math.max(0, order.quantity - 500) * order.itemPrice * 0.05;
const shipping = Math.min(basePrice * 0.1, 100);
return basePrice - quantityDiscount + shipping;
```

## Inline Variable

**When**: Variable name doesn't add meaning beyond the expression.

**Steps**:

1. Check variable is assigned only once
2. Replace all references with expression
3. Remove declaration

```typescript
// Before
const basePrice = order.basePrice;
return basePrice > 1000;

// After
return order.basePrice > 1000;
```

## Replace Temp with Query

**When**: Temporary variable holds calculation that could be a function.

**Steps**:

1. Extract calculation into function
2. Replace temp references with function call
3. Remove temp declaration

```typescript
// Before
function price(order: Order): number {
  const basePrice = order.quantity * order.itemPrice;
  if (basePrice > 1000) return basePrice * 0.95;
  return basePrice * 0.98;
}

// After
function price(order: Order): number {
  if (basePrice(order) > 1000) return basePrice(order) * 0.95;
  return basePrice(order) * 0.98;
}

function basePrice(order: Order): number {
  return order.quantity * order.itemPrice;
}
```
