# Load Generator Service

## Overview
The Load Generator service is not part of the actual "application logic." Instead, it is a synthetic traffic simulator meant to hammer the frontend with realistic user interactions.

* **Language:** Python (using Locust)
* **Port Exposed:** None (Runs endlessly in the background)

## How the Code Works
1. This service uses an open-source Python framework called **Locust**, which specializes in defining "user behaviors" rather than raw HTTP-bombing.
2. Inside `locustfile.py`, there is a `WebsiteUser` class configured to spawn fake users endlessly.
3. These fake users execute realistic traversal tasks:
   - They load the homepage.
   - They click on random product URLs.
   - They change the site currency arbitrarily.
   - They add items to their carts.
   - They completely execute the checkout pipeline.
4. When deployed in Kubernetes via `loadgenerator.yaml`, it targets `http://frontend:80` out of the box.

## Potential Trainer Questions

**Q: What is the primary purpose of deploying this inside a Kubernetes cluster?**
**A:** "The primary point is to demonstrate the power of **Horizontal Pod Autoscaling (HPA)**. With the load generator constantly spamming the frontend with computationally heavy cart checkouts, we can actually trigger real CPU limits. If we set an HPA threshold to 70% CPU, we can physically watch Kubernetes detect the spike and auto-provision 5, 10, or 20 additional frontend pods live, distributing the traffic flawlessly!"

**Q: Which services are under the most pressure when this runs?**
**A:** "Everything is hit, but the highest load falls heavily on the `Currency Service` (converting prices on every single catalog load), the `Frontend` (aggregating data), and now our new `Postgres Database` (handling continuous `INSERT` statements for random cart caching and order confirmations)."
