# API Refactorings

## Introduce Parameter Object

**When**: Same parameters travel together; group reveals deeper abstraction.

**Steps**:

1. Create class/type for grouped parameters
2. Add parameter of new type
3. Replace individual parameters
4. Consider moving behavior into new class

```typescript
// Before: Range parameters appear together everywhere
function amountInvoicedIn(start: Date, end: Date): number {}
function amountReceivedIn(start: Date, end: Date): number {}
function amountOverdueIn(start: Date, end: Date): number {}

// After: DateRange reveals domain concept
class DateRange {
  constructor(
    readonly start: Date,
    readonly end: Date
  ) {}

  contains(date: Date): boolean {
    return date >= this.start && date <= this.end;
  }
}

function amountInvoicedIn(range: DateRange): number {}
function amountReceivedIn(range: DateRange): number {}
function amountOverdueIn(range: DateRange): number {}
```

## Remove Flag Argument

**When**: Boolean/enum parameter changes function behavior; separate functions are clearer.

**Steps**:

1. Create explicit function for each flag value
2. Replace callers with explicit version
3. Remove original or leave as wrapper

```typescript
// Before: What does true mean?
setDimension(name, 10, true);

// After: Intent is clear
setWidth(10);
setHeight(10);

// Implementation
function setWidth(value: number): void {
  setDimension('width', value);
}

function setHeight(value: number): void {
  setDimension('height', value);
}
```

## Preserve Whole Object

**When**: Passing multiple values extracted from single object.

**Steps**:

1. Add parameter for whole object
2. Adjust function to extract values
3. Update callers to pass whole object
4. Consider if dependency is acceptable

```typescript
// Before: Extracting values just to pass them
const low = room.daysTempRange.low;
const high = room.daysTempRange.high;
if (plan.withinRange(low, high)) {
}

// After: Pass the whole object
if (plan.withinRange(room.daysTempRange)) {
}

class HeatingPlan {
  withinRange(range: TempRange): boolean {
    return range.low >= this._temperatureRange.low && range.high <= this._temperatureRange.high;
  }
}
```

## Replace Parameter with Query

**When**: Parameter can be determined from other parameters or context.

**Steps**:

1. Extract calculation if needed
2. Replace parameter references with calculation
3. Remove parameter from declaration and calls

```typescript
// Before: quantity can be derived
finalPrice(basePrice: number, discountLevel: number, quantity: number): number {
  return basePrice * (1 - this.discountFor(discountLevel, quantity));
}

// After: Calculate quantity internally
finalPrice(basePrice: number, discountLevel: number): number {
  const quantity = this.order.quantity;
  return basePrice * (1 - this.discountFor(discountLevel, quantity));
}
```

## Replace Query with Parameter

**When**: Need to reduce dependencies or make function more flexible.

**Steps**:

1. Extract calculation if needed
2. Add parameter to function
3. Replace internal query with parameter
4. Update callers to pass value

```typescript
// Before: Function depends on global thermostat
class HeatingPlan {
  get targetTemperature(): number {
    if (thermostat.selectedTemperature > this._max) return this._max;
    if (thermostat.selectedTemperature < this._min) return this._min;
    return thermostat.selectedTemperature;
  }
}

// After: Caller provides temperature
class HeatingPlan {
  targetTemperature(selectedTemp: number): number {
    if (selectedTemp > this._max) return this._max;
    if (selectedTemp < this._min) return this._min;
    return selectedTemp;
  }
}

// Caller
plan.targetTemperature(thermostat.selectedTemperature);
```

## Separate Query from Modifier

**When**: Function returns value AND has side effects.

**Steps**:

1. Copy function for query version
2. Remove side effects from query
3. Remove return from modifier
4. Replace callers: query for value, modifier for effect

```typescript
// Before: getTotalAndSendBill does two things
function getTotalAndSendBill(): number {
  const total = orders.reduce((sum, o) => sum + o.amount, 0);
  sendBill(total);
  return total;
}

// After: Separate concerns
function getTotal(): number {
  return orders.reduce((sum, o) => sum + o.amount, 0);
}

function sendBill(): void {
  sendBillEmail(getTotal());
}

// Caller
const total = getTotal();
sendBill();
```

## Remove Setting Method

**When**: Field should be set only at construction time.

**Steps**:

1. Check field is only set in constructor
2. Add to constructor parameters if needed
3. Remove setter method
4. Make field readonly

```typescript
// Before: id can be changed after creation
class Person {
  private _id: string = '';

  get id(): string {
    return this._id;
  }
  set id(value: string) {
    this._id = value;
  }
}

// After: Immutable after construction
class Person {
  constructor(private readonly _id: string) {}

  get id(): string {
    return this._id;
  }
}
```
