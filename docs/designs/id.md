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

### Additional requirements (if possible)

- When specifying a part of the ID, like Git commit hash, it can still identify the runbook if it is unique.
- Can rerun by `runn run [runbook ID]`

## Data structure

TODO

## Algorithm

TODO

## Alternatives considered

### Generate ID from `desc:` of runbook

No guarantee that `desc:` is unique, so other sources are needed for ID generation.

### Generate ID from `steps:` of runbook

If the `steps:` are the same, they can be considered the same runbook.

However, for fixes, when a step is modified or a new step is added, it becomes a different `steps:`, so it cannot be the source of ID generation.

### Generate ID from absolute path of runbook

Runbook file paths are unique on the file system.

And, in only a few cases do they change the file name or file path, either on rerun or when fixing a failure.

However, if the running environment is different, the absolute path of the runbook will be different.

### Generate ID from relative path of runbook

Different run paths will change the relative paths of runbooks.
