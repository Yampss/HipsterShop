# Recommendation Service

## Overview
The Recommendation Service powers the "You might also like..." section universally seen on e-commerce sites. It inspects what is currently in the user's cart and recommends correlated products.

* **Language:** Python
* **Port Exposed:** 8080 (gRPC)

## How the Code Works
1. The service receives a list of `product_ids` currently in the user's cart.
2. It makes a direct gRPC outbound call down to the `ProductCatalogService` to fetch the complete inventory list. Wait, a backend microservice calling another backend microservice? Yes!
3. Instead of implementing a complex machine learning tensor model, this demo uses a simple algorithm: it randomly samples 5 item IDs from the entire catalog, explicitly filters out any items that the user already has in their cart, and returns the resulting truncated list back up to the frontend UI.

## Potential Trainer Questions

**Q: Why is it written in Python? Isn't Python slow compared to Go?**
**A:** "Yes, Python is slower for raw network processing. However, almost all authentic AI, Machine Learning, and Recommendation engines in the industry (like TensorFlow and PyTorch) are natively supported by Python first. This service was explicitly written in Python to architecturally prepare it for a future swap-out to a real Machine Learning algorithm."

**Q: How does this service securely find the Product Catalog service?**
**A:** "Kubernetes internal CoreDNS. Look at the `recommendationservice.yaml` manifest. We inject an environment variable `PRODUCT_CATALOG_SERVICE_ADDR` with the value `productcatalogservice:3550`. Our Python code simply attempts a gRPC connection to that bare hostname. Kubernetes' internal DNS system magically catches the request and routes it to the correct internal IP address of the active product catalog pods. Total abstraction!"
