# Product Catalog Service

## Overview
The Product Catalog Service governs the retail inventory. It retrieves lists of products, searches products by keyword, and retrieves specific product details. We migrated this away from a local JSON file into a real MongoDB database.

* **Language:** Go (Golang)
* **Port Exposed:** 3550 (gRPC)

## How the Code Works
The main logic is in `catalog_loader.go`. 
1. When the container starts, it establishes a connection to the MongoDB replica set using the official Go MongoDB driver (`go.mongodb.org/mongo-driver`).
2. It queries the `shopdb.products` collection, fetching all product documents.
3. It decodes each BSON document into a Go struct (`mongoProduct`) that maps to the Protobuf `Product` message definition (handling prices, IDs, categories, and image paths).
4. The service then holds this catalog in memory and rapidly responds to gRPC calls from the Frontend (for rendering the homepage) and from the Checkout Service (for calculating checkout totals).

## Potential Trainer Questions

**Q: You migrated this to MongoDB. Where did the initial data come from?**
**A:** "In our `mongodb.yaml` manifest, we created a ConfigMap containing an initialization script. This script runs `db.products.insertMany()` to seed all the base inventory into the `shopdb.products` collection. The script is idempotent — it checks `countDocuments()` before seeding, so it only runs on first launch."

**Q: Why use the MongoDB Go driver instead of an ORM?**
**A:** "For this microservice, we only perform simple `Find()` queries to load the full product catalog into memory. An ORM would add unnecessary complexity and overhead. The official MongoDB Go driver gives us direct, efficient BSON decoding into Go structs with minimal abstraction."
