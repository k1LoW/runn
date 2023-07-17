# Runbook ID design doc

Authors: @k1low, @k2tzumi
Status: Draft

## Objective

This document describes the implementation of runbook ID.

## Backgroud

runn runs multiple runbooks.
When the run of one of multiple runbooks fails, there are the following use cases

- To identify the runbook/step that failed.
- To rerun the failed runbook. Rerun environment may be different (on local, on CI)
- To modify the failed runbook and rerun it.

The ID that identifies the runbook is useful in these use cases.

## Data structure

TODO

## Algorithm

TODO

## Alternatives considered

TODO
