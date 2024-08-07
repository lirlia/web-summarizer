# Web Summarizer

## Introduction

This project uses the Slack Websocket API to retrieve messages from Slack channels, extract URLs from those messages, and automatically summarize the content of those URLs.

And then, it sends the summary back to the Slack DM.

## How to use

1. create slack bot
1. update compose.yaml (add your keys)
1. run `docker-compose up -d`
