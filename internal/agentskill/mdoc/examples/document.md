---
mdoc: true
theme: plain
title: "Project Helios — Status Report"
author: "Platform Team"
tags: [report, status]
page:
  size: A4
  margin: 25mm 22mm 28mm 22mm
data:
  project: Helios
---

# {{.Data.project}} — Status Report

*Prepared by {{.Author}} on {{.System.Date}}.*

## Summary

Helios shipped the new ingestion pipeline this sprint. Throughput improved and
the backlog cleared.

## Metrics

| Metric       | Last sprint | This sprint |
| :----------- | ----------: | ----------: |
| Throughput/s |       1,200 |       3,400 |
| Error rate   |        2.1% |        0.4% |

The error rate now satisfies $\varepsilon < 0.5\%$, our SLO target:

$$
\varepsilon = \frac{\text{failed requests}}{\text{total requests}} < 0.005
$$

## Checklist

- [x] Ship ingestion v2
- [x] Backfill historical data
- [ ] Decommission the legacy path

## Notes

The legacy path stays online until Q4 for safety[^1].

[^1]: Rollback insurance while v2 soaks in production.
