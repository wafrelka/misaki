# Misaki

Tools for executing preset commands via the web interface.

<img src="./docs/screenshot.png" width="400">

## Architecture

```
misaki server ---> Amazon SQS ---> misaki executor ---> Slack (Incoming Webhooks)
```
