# Handling of the process environment

Authors: @k1low

Status: Accepted

## Objective

This document defines how runn reads the process environment, when the values it reads are fixed, and where runn is allowed to write to it.

## Background

runn reads the process environment in several places.

| Consumer | Where | How it reads |
| --- | --- | --- |
| `env.*` in expressions | `internal/store` | Snapshot held by the `Store` |
| `${VAR:-default}` expansion of the runbook YAML | `runbook.go` | `os.LookupEnv` at parse time |
| Child processes of the exec runner | `exec.go` | `os.Environ()` at step run time |
| runn's own settings (`RUNN_SCOPES`, `RUNN_RUN`, `RUNN_DEBUG`, ...) | `operator.go`, `cdp.go` | `os.Getenv` at `Load` time |
| Dependencies (proxy settings, DB drivers, cloud credentials, ...) | outside runn | Whatever the library does |

`env.*` is the consumer that needs rules. It is read on every expression evaluation, so the values it returns need a defined lifetime. runn also runs both as a CLI and as a Go test helper, so the process environment itself needs a defined owner.

## Rules

### `env` is a snapshot taken at the start of each run

The `Store` takes a snapshot of the process environment at the start of each run of a runbook, in `operator.runInternal`, right after the results of the previous run are cleared and before the runbook-level `if:` is evaluated. That snapshot is what `env.*` refers to for the rest of the run.

Consequences:

- Changes to the process environment made after `runn.New()` / `runn.Load()` but before `Run()` / `RunN()` are reflected. `t.Setenv()` in a Go test can be called at any point before the run.
- Changes made while the run is in progress are not reflected in that run. This includes `os.Setenv` from a `beforeFunc`, from a custom function registered with `runn.Func()`, or from any other goroutine. The `if:` condition, every step, and every included runbook see the same values.
- Root `loop:` iterations and `loadt` requests each start a new run, so each takes a new snapshot.

Secrets can refer to `env.*`, so the mask keywords of `secrets:` are re-registered whenever the snapshot is taken.

### Included runbooks share the snapshot of the parent

`include:` creates a nested operator with its own `Store`. The nested `Store` does not take a snapshot of its own. It receives the parent's snapshot in `operator.newNestedOperator`, so a runbook and everything it includes always agree on `env`, however deep the nesting.

Runbooks referenced by `needs:` are separate runs and take their own snapshot.

### runn as a library never writes to the process environment

Nothing in the `runn` package calls `os.Setenv`. When runn is used as a Go test helper, the caller owns the process environment and has `t.Setenv()` and `runn.Var()` to control what runbooks see. A library API that mutates the process environment would compete with `t.Setenv()` and would be invisible in the caller's code.

The `env` map exposed through the store is shared by every evaluation in the run. Consumers must treat it as read-only.

### The CLI `--env-file` option writes to the process environment, once, at startup

`--env-file` follows the semantics of docker's option of the same name. The file is applied to the environment of the process itself, as if the variables had been exported in the shell before invoking runn. It is implemented with `os.Setenv` in `internal/flags`, and it runs before any operator is created, so every consumer in the table above sees the same values, including runn's own `RUNN_*` settings and dependencies that read the environment directly.

This is deliberately a CLI concern. An operator-scoped option such as `runn.EnvFile(path)`, overlaying values without touching the process environment, was considered and rejected. It could reach only `env.*`, `${}` expansion and the exec runner, so the CLI option and the library option would have been two different features under one name. If a per-operator overlay is ever needed, it can be built on `Store.SetEnv()` without changing the rules above.

## Summary

- `env` is the process environment as of the start of the run, and it is immutable for the duration of the run.
- Included runbooks inherit the parent's snapshot.
- The `runn` package reads the process environment and never writes to it.
- `--env-file` is the only place runn writes to the process environment, and it happens once at CLI startup.
