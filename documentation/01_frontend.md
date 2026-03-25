# Frontend Service

## Overview
The Frontend service is the entry point for the entire application. It acts as both the web server that renders HTML to the user's browser, and as a lightweight API Gateway that aggregates data from all the backend microservices.

* **Language:** Go (Golang)
* **Port Exposed:** 8080
* **Kubernetes Service Type:** `LoadBalancer` (Exposes it to the public internet)

## How the Code Works
When a user opens the home page, the Go code executes an HTTP handler. 
1. It sends a gRPC request to the `ProductCatalogService` to get the list of items.
2. It sends a gRPC request to the `CurrencyService` to get exchange rates based on the user's cookie.
3. It sends a request to the `CartService` using a generated session ID to grab the user's cart count.
4. It compiles all this data and passes it into a Go HTML template (located in the `/templates` folder) to render the final web page.

Static assets (like the CSS, fonts, and the `.jpg` product images) are served directly by this service out of its `/static` directory.

## Potential Trainer Questions

**Q: How is the frontend exposed to the outside world?**
**A:** "In our Kubernetes `frontend.yaml`, we define a Kubernetes `Service` of type `LoadBalancer`. On AWS EC2/EKS, this automatically provisions an external Cloud Load Balancer that routes HTTP port 80 traffic down into our frontend pods' port 8080."

**Q: Does the frontend talk directly to the database?**
**A:** "No, and that is a crucial microservices principle. The frontend has absolutely no database credentials. It only talks to domain-specific services (like Cart and Catalog) via gRPC, and those services handle the secure database queries."
