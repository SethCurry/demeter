---
title: "Getting Started"
summary: ""
date: 2023-09-07T16:13:18+02:00
lastmod: 2023-09-07T16:13:18+02:00
draft: false
weight: 1
toc: true
---

Welcome to Demeter!

## A Quick Overview

```mermaid
architecture-beta
    group esp(server)[ESP32]
    group demeter(server)[Demeter]

    service api(server)[REST API] in demeter
    service db(disk)[SQLite] in demeter
    api:R --> L:db

    junction websocket in esp

    service camera(disk)[Camera] in esp
    camera:T -- B:websocket

    service thermometer(disk)[Thermometer] in esp
    thermometer:T -- B:websocket

    service hygrometer(disk)[Hygrometer] in esp
    hygrometer:T -- B:websocket

    websocket:T --> B:api

    align row api db
    align row thermometer camera hygrometer
    align column api websocket
```
