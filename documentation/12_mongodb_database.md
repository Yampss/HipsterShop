# MongoDB Database Service

## Overview
We migrated the database layer from a single PostgreSQL `Deployment` to a production-grade MongoDB `StatefulSet` with a dynamically configured Replica Set (`rs0`). Roles are determined at runtime via hostname ordinal extraction — no hardcoded replica count.

* **Database Engine:** MongoDB 7.0 (`mongo:7.0`)
* **Port Exposed:** 27017 (TCP)
* **Replica Set:** `rs0` (dynamic PRIMARY/SECONDARY assignment)
* **Architecture:** StatefulSet with Headless Service

## How the Setup Works

The `mongodb.yaml` manifest contains:

1. **The Secret (`mongodb-secret`):** Stores root username and password, referenced via `secretKeyRef`.

2. **Static PersistentVolumes (3×):** `hostPath` PVs provide 1Gi dedicated storage per pod via `volumeClaimTemplates`.

3. **Headless Service (`mongodb-headless`):** `clusterIP: None` gives each pod stable DNS: `mongodb-{N}.mongodb-headless`.

4. **ConfigMap (`mongodb-scripts`):** Two scripts using hostname ordinal extraction:
   - `init-mongo.sh` — Extracts ordinal from `$HOSTNAME` (`[[ $HOSTNAME =~ -([0-9]+)$ ]]`), writes it to a shared volume, generates the replica set keyfile.
   - `rs-configure.sh` — Reads the ordinal:
     - **Ordinal 0 (PRIMARY):** `rs.initiate()` with just itself, creates root user, seeds product data
     - **Ordinal N>0 (SECONDARY):** Waits for PRIMARY, then `rs.add(self)` to join the replica set

5. **StatefulSet (`mongodb`):**
   - **Init container** — runs `init-mongo.sh` to prepare config
   - **Main container** — `mongod --replSet rs0 --bind_ip_all`
   - **Sidecar container** — runs `rs-configure.sh` for RS setup
   - **Liveness/Readiness probes** — `mongosh --eval "db.adminCommand('ping')"`

6. **ClusterIP Service (`mongodb`):** Port 27017 for microservice access.

## Dynamic Ordinal-Based Configuration

```
Pod starts → init-mongo.sh extracts ordinal from $HOSTNAME
  │
  ├── ordinal == 0:  rs.initiate({self}) → create user → seed data
  └── ordinal > 0:   wait for PRIMARY → rs.add(self)
```

No script knows how many replicas exist. Scaling up (`kubectl scale`) automatically adds more SECONDARYs — no YAML changes needed.

## Connection String

```
mongodb://admin:<password>@mongodb-0.mongodb-headless:27017,mongodb-1.mongodb-headless:27017,mongodb-2.mongodb-headless:27017/shopdb?replicaSet=rs0&authSource=admin
```

## Potential Trainer Questions

**Q: Why did you switch from PostgreSQL to MongoDB?**
**A:** "MongoDB's document model naturally fits our use cases — cart data as BSON binary and products as semi-structured documents. The StatefulSet with replica set provides built-in HA and automatic failover."

**Q: How does the replica set configure itself without knowing the replica count?**
**A:** "Each pod extracts its ordinal index from its hostname using a regex. Pod-0 initiates the replica set with just itself. Any subsequent pod connects to pod-0 (the PRIMARY) and adds itself using `rs.add()`. This pattern scales to any number of replicas."

**Q: What happens if the PRIMARY crashes?**
**A:** "MongoDB automatically elects a new PRIMARY from the remaining SECONDARYs. The driver in our microservices discovers the new PRIMARY via the replica set connection string — zero manual intervention."
