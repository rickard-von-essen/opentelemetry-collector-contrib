# Confluent Cloud Receiver

| Status                   |                            |
| ------------------------ |----------------------------|
| Stability                | [alpha]                    |
| Supported pipeline types | metrics                    |
| Distributions            | [contrib]                  |

This receiver will scrape metrics from Confluent Cloud's Metric API v2.
See [Confluent Cloud Metrics API](https://api.telemetry.confluent.cloud/docs)

**TODO**

The receiver will watch the directory and read files. If a file is updated or added,
the receiver will read it in its entirety again.

Please note that there is no guarantee that exact field names will remain stable.
This intended for primarily for debugging Collector without setting up backends.

Supported pipeline types: metrics

## Getting Started

The following settings are required:

- `include`: set a glob path of files to include in data collection

Example:

```yaml
receivers:
  confluentcloud:
    TODO
```
