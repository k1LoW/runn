# Flow up to the runbook run by the operator.

Authors: @k1low

Status: Add continuously

## Objective

This document describes the flow up to the runbook run by operator/operatorN.

## Background

runn.operator and runn.operatorN are the entities that run their runbooks.

runn.operator operates the run of a single runbook.

runn.operatorN operates multiple runbooks (i.e., multiple operators) together.

## Flow

``` mermaid
flowchart TB
  runbook[A single runbook] --> New --Initialize operator--> operator.Run --Convert to operatorN using operator.toOperatorN--> operatorN.runN
  runbooks[Multiple runbooks] --> Load --Initialize operatorN--> operatorN.RunN --> operatorN.runN
  operatorN.runN --Run each runbooks--> operator.run
  operator.run --loop:--> operator.runInternal
  operator.run --> operator.runLoop
  operator.runLoop --> operator.runInternal
```
