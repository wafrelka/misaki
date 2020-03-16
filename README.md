# Misaki

Tools for executing preset commands via the web interface.

![screenshot.png](./docs/screenshot.png)

## Architecture

```
misaki server ---> Amazon SQS ---> misaki executor ---> Slack (Incoming Webhooks)
```
