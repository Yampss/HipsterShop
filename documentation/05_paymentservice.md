# Payment Service

## Overview
The Payment Service simulates charging a credit card for an order. It does not connect to any real external API (like Stripe or PayPal), but instead mocks the transaction response.

* **Language:** Node.js
* **Port Exposed:** 50051 (gRPC)

## How the Code Works
1. Runs a gRPC server defined in `server.js` using the `@grpc/grpc-js` library.
2. Accepts a `ChargeRequest` containing a `Money` amount and `CreditCardInfo`.
3. In `charge.js`, it performs rudimentary validation:
   - Throws an error if the card number is less than 16 digits.
   - Throws an error if the card is expired (month/year validation).
4. If validation passes, it forcefully accepts the mock transaction, generates a fake UUID `transaction_id`, and returns it immediately.

## Potential Trainer Questions

**Q: What is the purpose of this service if it doesn't actually talk to a payment gateway?**
**A:** "In a microservices architecture, a mock service like this allows frontend and backend developers to test their checkout orchestration workflows without accidentally incurring real credit card charges. It effectively tests the gRPC boundaries. Replacing this mock with a real Stripe integration would not require any changes to the Frontend or Checkout services, because the gRPC contract remains identical."

**Q: How does Node.js handle high concurrency compared to Go?**
**A:** "Node.js uses an asynchronous, single-threaded event loop. While not structurally parallel like Go routines, its non-blocking I/O allows it to handle tens of thousands of simultaneous gRPC network requests extremely efficiently without locking or waiting."
