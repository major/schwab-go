# OpenAPI specifications

This directory stores pretty-printed OpenAPI specifications captured from Schwab's developer portal for later comparisons against the client implementation and future Schwab API changes.

| File | Source capture | Portal file name | OpenAPI version | Title |
|---|---|---|---|---|
| `market_data.openapi.json` | `market_data.har`, entry 53 | `TraderApi-MDIS-03-21-2024(4).json` | 3.0.3 | Market Data |
| `trader_api.openapi.json` | `trader_api.har`, entry 20 | `TraderApi-Prod_05-11-2024.yaml` | 3.0.1 | Trader API - Account Access and User Preferences |

The source HAR files are intentionally ignored by git because browser captures can contain authorization headers, cookies, and other session data. Commit only the extracted specification JSON files unless a capture has been reviewed and scrubbed.
