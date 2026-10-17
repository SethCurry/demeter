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
usecase-beta
  direction LR
  actor User

  systemBoundary DemeterWeb
    WebAPI("HTTP API")
  end

  systemBoundary DemeterSensor
    Metrics("Metrics")
  end

  User --> WebAPI
  Metrics -- "websocket" --> WebAPI
```
