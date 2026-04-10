# Email Service Configuration & Testing Guide

This guide describes how to configure and test actual email sending from the `emailservice`. Previously, the service only received requests and wrote them to stdout or a Mongo database (dummy mode). It has now been updated to use standard `smtplib` to actually send emails.

## Configuration

To enable the email sending capability, you must provide your SMTP credentials to the service. This is done via the following environment variables:

- `SMTP_HOST`: The host address of your SMTP server (e.g., `smtp.gmail.com`, `smtp.sendgrid.net`). If this is not set, the service will fall back to "dummy mode."
- `SMTP_PORT`: The port for your SMTP server. (Default is `587`. Typically `587` for TLS or `465` for SSL).
- `SMTP_USER`: The username for authentication (e.g., your full Gmail address or SendGrid username).
- `SMTP_PASSWORD`: The password (or app-specific password/API key) for authentication.

### Important:
If you are using Gmail, you **must** use an **App Password**; regular account passwords will not work due to Google's security policies.

## How to Test Locally

1. **Set Environment Variables**:
   In your terminal, set the necessary environment variables before running the application:

   **For Windows (PowerShell):**
   ```powershell
   $env:SMTP_HOST="smtp.gmail.com"
   $env:SMTP_PORT="587"
   $env:SMTP_USER="your-email@gmail.com"
   $env:SMTP_PASSWORD="your-app-password"
   ```

   **For Linux/macOS:**
   ```bash
   export SMTP_HOST="smtp.gmail.com"
   export SMTP_PORT="587"
   export SMTP_USER="your-email@gmail.com"
   export SMTP_PASSWORD="your-app-password"
   ```

2. **Run the Email Service**:
   Run the Flask server from within the `src/emailservice` directory:
   ```bash
   pip install -r requirements.txt
   python email_server.py
   ```

3. **Send a Test Request**:
   You can mimic the `checkoutservice` and call the endpoint locally to see if it works. Use `curl` or Postman to send a POST request to the `/send-confirmation` endpoint.

   **Example `curl` Command:**
   ```bash
   curl -X POST http://localhost:8080/send-confirmation \
     -H "Content-Type: application/json" \
     -d '{
       "email": "receiver-email@example.com",
       "order": {
         "orderId": "12345-ABCDE",
         "shippingTrackingId": "TRACK-987654321",
         "shippingCost": {"currencyCode": "USD", "units": 5, "nanos": 0},
         "shippingAddress": {
           "streetAddress": "123 Main St",
           "city": "Sample City",
           "state": "CA",
           "country": "USA",
           "zipCode": "90210"
         },
         "items": [
           {
             "item": {"productId": "OLJCESPC7Z", "quantity": 1},
             "cost": {"currencyCode": "USD", "units": 19, "nanos": 990000000}
           }
         ]
       }
     }'
   ```

4. **Verify Email Delivery**:
   Check the inbox of the email address you specified in the `"email"` field of your POST request. You should see an order confirmation email rendered with HTML! Also, check the console output of the `email_server.py` terminal for success or error logs.

## Troubleshooting
- **Authentication Error**: Make sure you are using an app password (not a login password), and your user/password combination is correct.
- **Connection Refused**: Double-check the host and port numbers. It usually must be 587 or 465.
- **Empty or no Email Received**: Check the service output logs. If `SMTP_HOST` is omitted, the terminal will only show "A request to send order confirmation ... has been received" and no email will be attempted.
