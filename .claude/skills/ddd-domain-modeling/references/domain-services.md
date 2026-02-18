# Domain Services

Domain services contain business logic that doesn't naturally fit within a single entity or value object.

## Table of Contents

- [When to Use Domain Services](#when-to-use-domain-services)
- [Characteristics](#characteristics)
- [Patterns](#patterns)
- [Testing Domain Services](#testing-domain-services)

## When to Use Domain Services

Use a domain service when the operation:

- Involves multiple entities or aggregates
- Requires business logic that doesn't belong to any single entity
- Implements a domain concept that is a "verb" rather than a "noun"
- Needs to enforce cross-entity invariants

| Scenario                              | Solution                                    |
| ------------------------------------- | ------------------------------------------- |
| Calculate order total from line items | Entity method: `order.calculateTotal()`     |
| Check if user can afford a purchase   | Domain service: `PaymentEligibilityService` |
| Transfer funds between accounts       | Domain service: `FundsTransferService`      |
| Validate an email format              | Value object: `Email.create()`              |

## Characteristics

Domain services are:

- **Stateless**: No instance state, operate purely on inputs
- **Pure**: No side effects, same inputs → same outputs
- **Framework-free**: No database, HTTP, or infrastructure dependencies
- **Synchronous**: No async operations (those belong in application layer)

```typescript
// ✅ Good: Pure domain service
export class PricingService {
  calculateDiscount(items: LineItem[], customerTier: CustomerTier): Money {
    // Pure calculation logic
  }
}

// ❌ Bad: Has infrastructure concerns
export class PricingService {
  constructor(private db: Database) {} // Infrastructure dependency

  async calculateDiscount(customerId: string): Promise<Money> {
    const customer = await this.db.findCustomer(customerId); // Async
    // ...
  }
}
```

## Patterns

### Calculation Service

```typescript
// src/domain/services/PricingService.ts
import { Money } from '../value-objects/Money';
import type { LineItem } from '../entities/LineItem';
import type { Discount } from '../value-objects/Discount';

export class PricingService {
  calculateSubtotal(items: LineItem[]): Money {
    if (items.length === 0) {
      return Money.zero('USD');
    }

    return items.reduce(
      (total, item) => total.add(item.price.multiply(item.quantity)),
      Money.zero(items[0].price.currency)
    );
  }

  applyDiscounts(subtotal: Money, discounts: Discount[]): Money {
    let total = subtotal;

    // Sort by type: percentage discounts first, then fixed
    const sorted = [...discounts].sort((a, b) => (a.type === 'percentage' ? -1 : 1));

    for (const discount of sorted) {
      if (discount.type === 'percentage') {
        const reduction = total.multiply(discount.value / 100);
        total = total.subtract(reduction);
      } else {
        total = total.subtract(Money.of(discount.value, total.currency));
      }
    }

    // Never go below zero
    return total.amount < 0 ? Money.zero(total.currency) : total;
  }

  calculateTax(amount: Money, taxRate: number): Money {
    return amount.multiply(taxRate / 100);
  }
}
```

### Validation Service

```typescript
// src/domain/services/OrderValidationService.ts
import type { Order } from '../entities/Order';
import type { Customer } from '../entities/Customer';
import { Money } from '../value-objects/Money';

export interface ValidationResult {
  isValid: boolean;
  errors: string[];
}

export class OrderValidationService {
  validateOrder(order: Order, customer: Customer): ValidationResult {
    const errors: string[] = [];

    // Business rule: Customer must be active
    if (!customer.isActive) {
      errors.push('Customer account is not active');
    }

    // Business rule: Order value within customer's credit limit
    if (order.total.amount > customer.creditLimit.amount) {
      errors.push('Order exceeds customer credit limit');
    }

    // Business rule: Minimum order value
    const minimumOrder = Money.of(10, order.total.currency);
    if (order.total.amount < minimumOrder.amount) {
      errors.push('Order must be at least $10');
    }

    // Business rule: No more than 100 items per order
    if (order.itemCount > 100) {
      errors.push('Order cannot exceed 100 items');
    }

    return {
      isValid: errors.length === 0,
      errors,
    };
  }
}
```

### Policy Service

```typescript
// src/domain/services/ShippingPolicyService.ts
import type { Order } from '../entities/Order';
import type { Address } from '../value-objects/Address';
import { Money } from '../value-objects/Money';

export interface ShippingOption {
  name: string;
  cost: Money;
  estimatedDays: number;
}

export class ShippingPolicyService {
  private readonly FREE_SHIPPING_THRESHOLD = Money.of(50, 'USD');

  getAvailableOptions(order: Order, destination: Address): ShippingOption[] {
    const options: ShippingOption[] = [];
    const currency = order.total.currency;

    // Standard shipping always available
    options.push({
      name: 'Standard',
      cost: this.calculateStandardCost(order, destination),
      estimatedDays: this.getStandardDeliveryDays(destination),
    });

    // Express available for orders under 50lbs
    if (order.totalWeight < 50) {
      options.push({
        name: 'Express',
        cost: Money.of(15.99, currency),
        estimatedDays: 2,
      });
    }

    // Overnight for domestic only
    if (destination.country === 'US') {
      options.push({
        name: 'Overnight',
        cost: Money.of(29.99, currency),
        estimatedDays: 1,
      });
    }

    return options;
  }

  private calculateStandardCost(order: Order, destination: Address): Money {
    // Free shipping over threshold
    if (order.total.amount >= this.FREE_SHIPPING_THRESHOLD.amount) {
      return Money.zero(order.total.currency);
    }

    // Base cost + weight surcharge
    const baseCost = destination.country === 'US' ? 5.99 : 14.99;
    const weightSurcharge = Math.max(0, (order.totalWeight - 5) * 0.5);

    return Money.of(baseCost + weightSurcharge, order.total.currency);
  }

  private getStandardDeliveryDays(destination: Address): number {
    return destination.country === 'US' ? 5 : 14;
  }
}
```

### Allocation/Distribution Service

```typescript
// src/domain/services/InventoryAllocationService.ts
import type { OrderItem } from '../entities/OrderItem';
import type { WarehouseStock } from '../entities/WarehouseStock';

export interface AllocationResult {
  allocations: Map<string, { warehouseId: string; quantity: number }[]>;
  unallocated: { productId: string; quantity: number }[];
}

export class InventoryAllocationService {
  allocateStock(items: OrderItem[], warehouseStocks: WarehouseStock[]): AllocationResult {
    const allocations = new Map<string, { warehouseId: string; quantity: number }[]>();
    const unallocated: { productId: string; quantity: number }[] = [];

    for (const item of items) {
      const productAllocations: { warehouseId: string; quantity: number }[] = [];
      let remaining = item.quantity;

      // Sort warehouses by available stock (highest first)
      const sortedStocks = warehouseStocks
        .filter((ws) => ws.productId === item.productId)
        .sort((a, b) => b.available - a.available);

      for (const stock of sortedStocks) {
        if (remaining <= 0) break;

        const allocate = Math.min(remaining, stock.available);
        if (allocate > 0) {
          productAllocations.push({
            warehouseId: stock.warehouseId,
            quantity: allocate,
          });
          remaining -= allocate;
        }
      }

      if (productAllocations.length > 0) {
        allocations.set(item.productId, productAllocations);
      }

      if (remaining > 0) {
        unallocated.push({ productId: item.productId, quantity: remaining });
      }
    }

    return { allocations, unallocated };
  }
}
```

## Testing Domain Services

Domain services are pure functions—test without mocks:

```typescript
describe('PricingService', () => {
  let service: PricingService;

  beforeEach(() => {
    service = new PricingService();
  });

  describe('calculateSubtotal', () => {
    it('sums all line items', () => {
      const items = [
        createLineItem({ price: Money.of(10, 'USD'), quantity: 2 }),
        createLineItem({ price: Money.of(5, 'USD'), quantity: 3 }),
      ];

      const result = service.calculateSubtotal(items);

      expect(result.amount).toBe(35); // (10*2) + (5*3)
    });

    it('returns zero for empty items', () => {
      const result = service.calculateSubtotal([]);
      expect(result.amount).toBe(0);
    });
  });

  describe('applyDiscounts', () => {
    it('applies percentage discount', () => {
      const subtotal = Money.of(100, 'USD');
      const discounts = [{ type: 'percentage' as const, value: 10 }];

      const result = service.applyDiscounts(subtotal, discounts);

      expect(result.amount).toBe(90);
    });

    it('applies fixed discount', () => {
      const subtotal = Money.of(100, 'USD');
      const discounts = [{ type: 'fixed' as const, value: 15 }];

      const result = service.applyDiscounts(subtotal, discounts);

      expect(result.amount).toBe(85);
    });

    it('applies percentage before fixed', () => {
      const subtotal = Money.of(100, 'USD');
      const discounts = [
        { type: 'fixed' as const, value: 10 },
        { type: 'percentage' as const, value: 10 },
      ];

      const result = service.applyDiscounts(subtotal, discounts);

      // 100 - 10% = 90, then 90 - 10 = 80
      expect(result.amount).toBe(80);
    });

    it('never returns negative', () => {
      const subtotal = Money.of(10, 'USD');
      const discounts = [{ type: 'fixed' as const, value: 50 }];

      const result = service.applyDiscounts(subtotal, discounts);

      expect(result.amount).toBe(0);
    });
  });
});

// Test helper
function createLineItem(props: { price: Money; quantity: number }): LineItem {
  return LineItem.create({
    productId: 'prod-1',
    price: props.price,
    quantity: props.quantity,
  });
}
```

### Integration with Use Cases

```typescript
// src/application/use-cases/PlaceOrder.ts
import { PricingService } from '@domain/services/PricingService';
import { ShippingPolicyService } from '@domain/services/ShippingPolicyService';
import { OrderValidationService } from '@domain/services/OrderValidationService';

export class PlaceOrder {
  constructor(
    private orderRepository: OrderRepository,
    private customerRepository: CustomerRepository,
    private pricingService: PricingService,
    private shippingService: ShippingPolicyService,
    private validationService: OrderValidationService
  ) {}

  async execute(request: PlaceOrderRequest): Promise<PlaceOrderResult> {
    const customer = await this.customerRepository.findById(request.customerId);
    if (!customer) throw new Error('Customer not found');

    // Use domain services for business logic
    const subtotal = this.pricingService.calculateSubtotal(request.items);
    const afterDiscounts = this.pricingService.applyDiscounts(subtotal, request.discounts);
    const shipping = this.shippingService.getAvailableOptions(
      { total: afterDiscounts, items: request.items },
      request.shippingAddress
    )[0]; // Use first option

    const order = Order.create({
      customerId: customer.id,
      items: request.items,
      subtotal,
      discounts: afterDiscounts,
      shipping: shipping.cost,
      total: afterDiscounts.add(shipping.cost),
    });

    // Validate using domain service
    const validation = this.validationService.validateOrder(order, customer);
    if (!validation.isValid) {
      return { success: false, errors: validation.errors };
    }

    await this.orderRepository.save(order);
    return { success: true, orderId: order.id };
  }
}
```
