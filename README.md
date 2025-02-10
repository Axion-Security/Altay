# Altay

![banner](https://github.com/user-attachments/assets/8999713a-df14-40e0-bb49-61d6e787c426)

**Altay** is a proof-of-concept exploit tool designed to demonstrate how vulnerabilities in browser extensions can be leveraged to capture and log visited URLs. By establishing a WebSocket connection, Altay communicates with a compromised browser extension (Altay BEX: Browser Extension) to receive real-time data from the target browser.

> **WARNING:**  
> **This software is for educational and research purposes only.**  
> Unauthorized use or deployment against systems without explicit consent is illegal and unethical. The author assumes no responsibility for any misuse or damage resulting from this tool. Always obtain proper authorization before conducting any testing.

## Potential Use Cases

Altay demonstrates how a compromised browser extension can be used to collect valuable reconnaissance data. In controlled, ethical environments and with proper authorization, similar techniques might be applied for:

- **Google Dork Parsing:**  
  Aggregated URLs can be used to refine search queries (Google dorks) aimed at identifying vulnerable targets or sensitive data exposed online.

- **Scanning Private/Unknown Forums:**  
  Collected URLs may assist in scanning private or lesser-known forums for vulnerabilities. In an authorized context, this data can help compile databases of potential exposures for further analysis.

## Features

- **Real-Time Communication:**  
  Maintains a persistent WebSocket connection with a compromised browser extension.

- **URL Logging:**  
  Captures and records visited URLs from the target browser, saving them to a designated directory.

- **Client Identification:**  
  Extracts the client’s IP address solely from HTTP headers (e.g., `X-Real-IP` and `X-Forwarded-For`).

- **Directory Management:**  
  Automatically creates and renames directories (with safe formatting) based on the browser's identification and IP address to store log files.

## How It Works

1. **Connection Establishment:**  
   The malicious browser extension connects to Altay Server via the `/ws` WebSocket endpoint.

2. **Client Identification:**  
   Upon connection, Altay retrieves the client’s IP address exclusively from the HTTP headers.

3. **Message Handling:**  
   - When the extension sends a message with the format `BrowserConnected|<BrowserName>`, Altay Server updates the client's details and renames the corresponding log directory.
   - When the extension sends `VisitedURL:<URL>`, Altay Server logs the URL by appending it to a file within the client’s directory.
