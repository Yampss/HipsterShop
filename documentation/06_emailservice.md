# Email Service

## Overview
The Email Service simulates sending an order confirmation email to the user when they successfully complete checkout. 

* **Language:** Python
* **Port Exposed:** 8080 (gRPC)

## How the Code Works
1. Runs a Python gRPC server using `grpcio`.
2. Receives a `SendOrderConfirmationRequest` carrying an email string and the full `OrderResult`.
3. The server takes the items, prices, and shipping tracking ID, aggregates them into a mock "Receipt" payload.
4. Instead of actually dispatching traffic to an SMTP server (like SendGrid or AWS SES), it simply prints the formatted email to `stdout` via `logger.py` and returns a successful response to the Checkout service.

## Potential Trainer Questions

**Q: If we wanted to make this send real emails, how hard would it be?**
**A:** "Extremely easy. We would just add a standard Python library like `smtplib` or an external API SDK like `boto3` for AWS SES into the `email_server.py`. The beauty of microservices is that we wouldn't need to rebuild or touch the other 10 containers; we just push an updated version of the `emailservice` container."

**Q: How do you monitor these fake emails since they just print to the console?**
**A:** "In a real Kubernetes environment, any logs printed to `stdout` inside a container are automatically captured by the Kubernetes node. We would use a centralized logging aggregator like Fluentd or Datadog to scrape those logs, allowing developers to see the mock emails globally without entering the individual pods."
