# Ad Service

## Overview
The Ad Service is designed to serve hyper-contextual advertisements randomly on the user's screen based on the content of the products they are currently viewing.

* **Language:** Java (Spring Boot)
* **Port Exposed:** 9555 (gRPC)

## How the Code Works
1. Runs inside a stripped-down Alpine Linux container with OpenJDK.
2. The core logic inside `AdService.java` relies heavily on an internal map representing an "ad inventory network."
3. When the user visits the `/product/{id}` route on the frontend, the frontend parses the category of the item (e.g., "camera", "vintage", "photography").
4. The frontend calls the Ad Service with an `AdRequest` containing those exact context keys.
5. The Java Ad Service rapidly iterates through its inventory network, fetching all advertisements that match any of the provided categories. It then returns a maximum of 2 distinct `Ad` objects to the frontend to render dynamically under the "You May Also Like" banner.

## Potential Trainer Questions

**Q: Why was Java chosen for this specific service?**
**A:** "In a real advertising ecosystem (like Google Ads or Meta Ads), ad bidding engines are staggeringly complex, highly-concurrent, mathematics-heavy environments. The JVM (Java Virtual Machine) has possessed unparalleled, battle-tested garbage collection and multi-threading capabilities for decades. For an enterprise handling millions of concurrent financial bids, Java is often the industry standard."

**Q: Can you explain how the frontend knows what to display from an Ad object?**
**A:** "The `Ad` protobuf message strictly returns two strings: a `text` value (which is what the user reads), and a `redirect_url` value (which the frontend embeds into the `href` attribute of an `<a>` link). The frontend trusts the Ad Service implicitly, rendering the raw data securely."
