# Product Catalog Service

## Overview
The Product Catalog Service governs the retail inventory. It retrieves lists of products, searches products by keyword, and retrieves specific product details. We migrated this away from a local JSON file into a real PostgreSQL database.

* **Language:** Go (Golang)
* **Port Exposed:** 3550 (gRPC)

## How the Code Works
The main logic is in `catalog_loader.go`. 
1. When the container starts, it establishes a connection pool (`pgxpool`) to the PostgreSQL server.
2. It executes a `SELECT` query to pull all inventory from the `products` table.
3. It iterates over the SQL rows, scanning the data into Go structs that map to the Protobuf `Product` message definition (handling prices, IDs, categories, and image paths).
4. The service then holds this catalog in memory and rapidly responds to gRPC calls from the Frontend (for rendering the homepage) and from the Checkout Service (for calculating checkout totals).

## Potential Trainer Questions

**Q: You migrated this to PostgreSQL. Where did the initial data come from?**
**A:** "In our `postgres.yaml` manifest, we created a Kubernetes `ConfigMap` containing an `init.sql` script with `INSERT` statements holding all the base inventory. We mounted this ConfigMap into the PostgreSQL container at `/docker-entrypoint-initdb.d/`. Official Postgres Docker images automatically execute any SQL files in this directory on first launch, seamlessly seeding our database!"

**Q: Why use a connection pool in Go?**
**A:** "Establishing a new TCP connection to a database for every single user request is extremely slow and resource-heavy. By utilizing Go's `pgxpool`, the microservice maintains a pool of persistent, open connections to Postgres, drastically reducing latency and preventing the database from crashing under high traffic."
