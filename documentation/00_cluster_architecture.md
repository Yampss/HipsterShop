# Cluster & Architecture Overview

## Getting the Big Picture
The "Online Boutique" is a cloud-native microservices application originally designed by Google to demonstrate Kubernetes best practices. The application consists of 11 stateless microservices written in 5 different programming languages (Go, Python, Java, C#, and Node.js), and one stateful database (MongoDB).

## Key Architectural Concepts to Mention During Presentation

1. **gRPC Protocol for Internal Communication**
   Except for the `frontend` which exposes an HTTP server for user browsers, all backend microservices communicate using **gRPC** over HTTP/2. 
   * **Why it matters:** gRPC relies on strict `.proto` (Protocol Buffer) contracts. It is highly compressed, strongly typed, and significantly faster than standard REST/JSON APIs, making it ideal for low-latency internal microservice communication.

2. **Decoupling and Language Diversity**
   The architecture explicitly uses different languages (Polyglot) to prove that in a Kubernetes environment, the underlying technology stack of a container does not matter. As long as a container speaks gRPC and exposes ports correctly, standard Docker images unify the deployment model.

3. **Stateful vs Stateless (Important!)**
   * **Stateless:** All 11 microservices are completely stateless. This means if a pod dies, Kubernetes can spin up a replacement instantly without any data loss. They can autoscale infinitely horizontally.
   * **Stateful:** The **MongoDB** database is the only stateful part of the cluster. We use a Kubernetes **StatefulSet** with a 3-member Replica Set (`rs0`) backed by **Persistent Volumes (PV)** and **Persistent Volume Claims (PVC)**. Each pod gets stable network identity (`mongodb-0`, `mongodb-1`, `mongodb-2`) and dedicated storage. If a pod crashes, its storage persists and is reattached automatically. The replica set provides automatic failover — if the PRIMARY dies, a SECONDARY is elected as the new PRIMARY.

4. **Self-Healing through Probes**
   Every service manifest features `livenessProbe` and `readinessProbe` blocks calling the `grpc.health.v1.Health` standard. This hands the responsibility of monitoring application health directly to the Kubernetes control plane.

## Questions Trainers Will Likely Ask

**Q: If your services are in different languages, how do they know how to talk to each other?**
**A:** "We use gRPC and Protocol Buffers. A developer writes a `demo.proto` file that defines the messages and API endpoints. The protobuf compiler (`protoc`) then automatically generates native networking code for Go, Python, Java, C#, and Node.js, ensuring guaranteed type-safety and contract enforcement."

**Q: What happens if a backend service crashes while a user is on the site?**
**A:** "Because our deployments use Kubernetes `LivenessProbes`, the cluster's Kubelet will immediately detect the failure and restart the pod. Since the services are stateless, traffic will briefly failover or retry, and once the new pod passes its `ReadinessProbe`, the cluster will seamlessly route traffic back to it without human intervention."

**Q: How does data survive if a node crashes?**
**A:** "We implemented a MongoDB StatefulSet with Persistent Volume Claims. Each replica pod has its own dedicated persistent storage. If a pod or node crashes, the persistent disk remains intact and Kubernetes remounts it when the pod restarts. Additionally, MongoDB's replica set ensures data is replicated across all 3 members, so even if one node is lost entirely, two copies of the data remain."
