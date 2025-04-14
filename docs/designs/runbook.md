# Runbook design doc

Authors: @k1low

Status: Add continuously

## Objective

This document lists the properties of the runbook.

## Backgroud

We define a runbook and runn runs operations using it.

runn developers need to have a common understanding of the properties of the runbook.

## Properties of the runbook

- The runbook is the minimum unit of running of runn. For example, it is not assumed to have features that run only certain steps in a runbook.
- The running of the steps in a runbook is always sequential. Concurrent running is not assumed.
- All steps that have started running will stop when that runbook is completed (e.g. `exec.background:`).
    - Steps marked with `defer:` have not started running, so run execution is delayed until completion of the parent runbook.
