# PostgreSQL Database Service

## Overview
We migrated this architecture to utilize a robust, stateful relational database to persist the `Cart Service` sessions and hold the `Product Catalog` inventory. This completely eliminated the original fragile, stateless implementations.

* **Database Engine:** PostgreSQL (postgres:15-alpine)
* **Port Exposed:** 5432 (TCP)

## How the Setup Works
The `postgres.yaml` manifest is radically different from the stateless microservices:
1. **The Secret:** We generated a Kubernetes `Secret` (`postgres-secret`) to store the raw, encoded password securely, rather than exposing it directly in a yaml environment variable. This allows the backend services to reference it by a secure key value lookup.
2. **The Persistent Volume Claim (PVC):** The `postgres-pvc` requests 1 Gigabyte of persistent network storage from the underlying node (e.g., an AWS EBS volume).
3. **The Deployment:** We mount that PVC directly to the Postgres storage directory (`/var/lib/postgresql/data`). This ensures database tables physically survive any container crashes.
4. **The ConfigMap (Initialization):** We created a `ConfigMap` called `postgres-init` holding our raw SQL `CREATE TABLE` and `INSERT` instructions. This is mounted to `/docker-entrypoint-initdb.d/`, instructing the Postgres container to natively execute the SQL schema creation to build our tables and instantly seed them with store inventory.

## Potential Trainer Questions

**Q: This project claims that microservices are supposed to be stateless. Isn't deploying a database inside Kubernetes an anti-pattern?**
**A:** "Great question! Storing state within Kubernetes *used* to be considered an anti-pattern years ago, but with the invention of the Container Storage Interface (CSI) and vastly improved StatefulSets/PVC mechanics, running databases inside clusters is extremely common in modern cloud environments. That being said, at a massive enterprise scale (like Netflix or Uber), it's still best practice to offload database responsibilities to a managed cloud service like AWS RDS or Google Cloud SQL, and only deploy the stateless application code in Kubernetes to ensure maximum performance."

**Q: Wait, why isn't this deployed as a `StatefulSet`? The YAML clearly says `Deployment`.**
**A:** "In a fully production scenario requiring active leader/follower replication across 5+ database nodes, we absolutely must use a `StatefulSet`. However, for this demo presentation containing a standalone, singular primary Postgres instance holding straightforward product caching data, a standard `Deployment` bound to a single PVC handles the state perfectly fine and is far easier to orchestrate than configuring replication election logic."
