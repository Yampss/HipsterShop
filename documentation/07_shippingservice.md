# Shipping Service

## Overview
The Shipping Service performs two critical functions:
1. Calculating shipping cost estimates based on the contents of the user's cart.
2. Generating a mock tracking identifier when an order ships.

* **Language:** Go (Golang)
* **Port Exposed:** 50051 (gRPC)

## How the Code Works
The core logic resides in `quote.go` and `tracker.go`.
1. For quotes, it receives the items in the cart (a list of `CartItem` from `demo.proto`), counts the total number of items, and arbitrarily calculates a shipping cost (e.g., $8.99 plus some pseudo-randomized cents). It returns this cost to the Checkout service.
2. For shipping an order, it generates a random alphanumeric string with an 18-character length using a specific seed pool, logs that the items have been dispatched to the provided address, and returns the fake tracking number.

## Potential Trainer Questions

**Q: Why doesn't the Frontend call the Shipping Service directly for a quote?**
**A:** "In our architecture, the Frontend doesn't calculate final totals directly to prevent tampering. When a user requests a quote, the Frontend calls the orchestrator (Checkout Service), which fetches the shipping variables, ensuring the backend establishes a source of truth for all pricing before allowing the user to proceed."

**Q: How does Go's performance help here?**
**A:** "Go compiles directly to strict machine code binaries, unlike Node.js or Python which execute through interpreters. This makes Go microservices incredibly fast, drastically reducing latency in computationally heavy tasks like generating millions of random UUID trackers under massive load."
