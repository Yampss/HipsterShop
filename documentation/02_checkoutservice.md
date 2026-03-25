# Checkout Service

## Overview
The Checkout Service acts as the "orchestrator" for the most complex transaction in the application: completing a purchase. 

* **Language:** Go (Golang)
* **Port Exposed:** 5050 (gRPC)

## How the Code Works
When the user clicks "Place Order" on the frontend web page, the Frontend sends a massive `PlaceOrderRequest` gRPC call to the Checkout Service containing the user's address and credit card details.

The Checkout Service executes the following workflow synchronously:
1. Calls `CartService` to retrieve the user's shopping cart items.
2. Calls `ProductCatalogService` to verify the ID and price of those items.
3. Calls `CurrencyService` to convert all prices to the user's chosen currency.
4. Calls `ShippingService` with the address to generate a mock tracking number.
5. Calls `PaymentService` with the total amount and credit card details to simulate a charge.
6. Calls `CartService` again to EMPTY the user's cart.
7. Calls `EmailService` to dispatch a fake receipt.
8. Returns an `OrderResult` to the Frontend.

## Potential Trainer Questions

**Q: What is orchestration in microservices, and how is it used here?**
**A:** "Orchestration is when one central controller service coordinates interactions between multiple independent services. The Checkout Service does not know *how* to charge a credit card or *how* to send an email; it just orchestrates the flow of data between the Payment, Shipping, Cart, and Email microservices to fulfill a complex business goal."

**Q: What happens if the Payment Service crashes while the Checkout service is running this sequence?**
**A:** "Because gRPC supports deadlines and built-in error handling, the Checkout service will catch the failure, abort the transaction, and return a clean error to the frontend. It prevents scenarios like emptying a cart without successfully charging the card."
