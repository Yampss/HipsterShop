# Currency Service

## Overview
The Currency Service handles live money conversions (e.g., USD to EUR, JPY, GBP) allowing the frontend to dynamically display localized pricing based on the user's selected preference.

* **Language:** Node.js
* **Port Exposed:** 7000 (gRPC)

## How the Code Works
1. When the container starts, it reads a heavily-populated internal json file containing hardcoded exchange rate multipliers (based on real European Central Bank data).
2. The core logic inside `server.js` exposes a `SupportedCurrencies` endpoint.
3. Once the frontend receives a base `Money` object and a target currency code, it sends it via gRPC. 
4. The service divides the base `units` and `nanos` by the base rate, then multiplies it against the target currency conversion rate, rounds to the nearest cent, reconstructs a `Money` protobuf payload, and sends it back.

## Potential Trainer Questions

**Q: This service has the highest Queries-Per-Second (QPS) in the entire cluster. Why?**
**A:** "Every time a user loads the homepage, the Frontend has to render 9 products. If the user changed their currency to EUR, the Frontend must make 9 separate rapid-fire gRPC calls to the Currency Service just to render a single page. If we have 1,000 users browsing, the Currency Service gets hit 9,000 times per second! This perfectly demonstrates why gRPC (with its zero-latency HTTP/2 multiplexing) is vastly superior to REST for internal inter-service chitchat."

**Q: If we wanted to fetch live currency rates daily, how would we architect it?**
**A:** "We could implement a Kubernetes `CronJob` that spins up once a day, hits an external Bank API, and writes the latest conversion multipliers into a shared Postgres table or Redis cache. We'd then update the Currency Service Node.js code to pull the latest rate from the database instead of the local hardcoded JSON file."
