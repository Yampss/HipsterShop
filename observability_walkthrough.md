# How Your New Observability Stack Works (EC2 NodePort Setup)

Congratulations! Because the `loadgenerator` pod is aggressively simulating dozens of users constantly clicking through the store and adding items to their carts, your cluster is generating extremely rich data right now.

Since you are running Kubeadm on Master/Worker EC2 instances, we will use **NodePorts (30000-32767)** to expose the dashboards directly to the public internet, rather than using `kubectl port-forward` (which only binds to the Master Node's local CLI).

Here is the exact guide on how this data flows internally, which commands to run to intercept it, and what you should look for.

---

## 1. OpenTelemetry Distributed Tracing (Jaeger) 🕵️

### How it internally works:
Every time a user visits the Frontend, they are assigned a unique 16-byte `trace_id`. When the Frontend makes a gRPC call to the Cart Service, it attaches that `trace_id` as metadata in the request headers (often called B3 or W3C Trace Context). 

Because we injected `ENABLE_TRACING=1` and `COLLECTOR_SERVICE_ADDR` into your YAMLs, every microservice now executes a background thread that silently beams these trace timing records via OTLP (OpenTelemetry Protocol on port 4317) to the Jaeger Collector pod residing in the `observability` namespace. 

Jaeger catches these traces, stitches them together using the `trace_id`, and visualizes the exact timeline of requests traversing your system.

### Commands to Run:
First, expose the Jaeger UI to the internet by converting its service to a NodePort:
```bash
kubectl patch svc jaeger -n observability -p '{"spec": {"type": "NodePort"}}'
```

Next, find out exactly which randomly assigned port (30000-32767) Kubernetes gave you:
```bash
kubectl get svc jaeger -n observability
```
Look at the `PORT(S)` column. You will see something like `16686:31456/TCP`. 

### Where to Look & What to Do:
1. Open your browser and navigate to `http://<ANY-EC2-PUBLIC-IP>:31456` (Replace `31456` with the exact 5-digit port you got above).
2. Under "Service", select **`frontend`** from the dropdown menu.
3. Click the blue **Find Traces** button.
4. You will see a list of recent user interactions. Click one that has **8+ Spans** (like `/cart/checkout`).
5. **Observe**: You will visually see exactly how the `frontend` calls the `checkoutservice`, which in turn talks to `paymentservice`, `emailservice`, etc. If any specific service is slow, the timeline bar for that service will stretch out loudly, allowing you to instantly identify bottlenecks!

---

## 2. Metrics & Cluster Monitoring (Prometheus & Grafana) 📊

### How it internally works:
While Tracing follows a *single* user request, Metrics observes the *overall* system heartbeat (CPU usage, memory footprint, request rates, 500 errors). 

The `kube-prometheus-stack` you installed automatically deployed a Prometheus Operator. This Operator is constantly scanning your cluster looking for Pods measuring and exposing metrics (usually hitting HTTP `/metrics` endpoints). Prometheus scrapes these endpoints every 15 seconds, stores the data in a fast time-series database, and Grafana is used to visually present those metrics through customizable dashboards.

### Commands to Run:
First, expose Grafana to the internet by converting its service to a NodePort:
```bash
kubectl patch svc monitoring-grafana -n monitoring -p '{"spec": {"type": "NodePort"}}'
```

Find out which port Grafana was assigned:
```bash
kubectl get svc monitoring-grafana -n monitoring
```
Look at the `PORT(S)` column. You will see something like `80:32109/TCP`.

### Where to Look & What to Do:
1. Open your browser to `http://<ANY-EC2-PUBLIC-IP>:32109` (using your unique 5-digit port from above).
2. **Login Credentials**:
   - Username: `admin`
   - Password: By default, the `kube-prometheus-stack` generates a secure random password. You can retrieve it by running this command on your terminal:
     ```bash
     kubectl get secret -n monitoring monitoring-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
     ```
3. Hover over the "Dashboards" icon (four squares) in the left sidebar and click **Browse**.
4. You will see dozens of pre-configured dashboards that Prometheus Operator built for you automatically.
5. Click on **Kubernetes / Compute Resources / Pod** or **Kubernetes / Compute Resources / Cluster**.
6. At the top, filter the namespace to `default`.
7. **Observe**: You will see beautiful, real-time graphs showing exactly how much RAM and CPU your Java `adservice` or C# `cartservice` is legally consuming inside the cluster over time.

---

## Summary of the Difference

If your users suddenly complain: *"The shop is slow!"*

1. You look at **Grafana Dashboards** (Prometheus) first: You see that CPU utilization on the `shippingservice` is spiked at 100%.
2. You look at **Jaeger Traces** (OpenTelemetry) second: You find a trace taking 5 seconds, drill into it, and realize the exact SQL query or function inside `shippingservice` that caused the delay.
