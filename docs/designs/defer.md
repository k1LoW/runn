# Behavior of `defer:`

Authors: @k1low

Status: Add continuously

## Objective

This document describes the `defer:` behavior in runn.

## Backgroud

Allow post-processing of the runbook to be described using `defer:`

## Behavior of `defer:`

As the name suggests, `defer:` is inspired by the [defer](https://go.dev/blog/defer-panic-and-recover) of the Go programming language.

Behavior is also similar to the defer of the Go programming language, but some differences.

The order of run of steps not marked `defer:` are as follows.

```yaml
# main.yml
steps:
  - desc: step 1
    test: true
  - desc: step 2
    test: true
  - desc: step 3
    test: true
  - desc: step 4
    include:
      path: include.yml
  - desc: step 5
    test: true
```

```yaml
# include.yml
steps:
  - desc: included step 1
    test: true
  - desc: included step 2
    test: true
  - desc: included step 3
    test: true
```

``` mermaid
flowchart TB
  subgraph "(main.yml)"
    A[step 1]
    B[step 2]
    C[step 3]
    G[step 5]
  end

  subgraph "step 4 (include.yml)"
    D[included step 1]
    E[included step 2]
    F[included step 3]
  end
  
  A --> B
  B --> C
  C --> D
  D --> E
  E --> F
  F --> G
```

The steps with `defer:` are run as follows.

```yaml
# main.yml
steps:
  - desc: step 1
    test: true
  - desc: step 2
    defer: true
    test: true
  - desc: step 3
    defer: true
    test: true
  - desc: step 4
    include:
      path: include.yml
  - desc: step 5
    test: true
```

```yaml
# include.yml
steps:
  - desc: included step 1
    test: true
  - desc: included step 2
    defer: true
    test: true
  - desc: included step 3
    test: true
```

``` mermaid
flowchart TB
  subgraph "(main.yml)"
    A[step 1]
    B["step 2 (defer: true)"]
    C["step 3 (defer: true)"]
    G[step 5]
  end

  subgraph "step 4 (include.yml)"
    D[included step 1]
    E["included step 2 (defer: true)"]
    F[included step 3]
  end
  
  A --> D
  D --> F
  F --> G
  G --> E
  E --> C
  C --> B
```

The step marked `defer` behaves as follows.

- If `defer: true` is set, run of the step is deferred until finish of the runbook.
- Steps marked with `defer` are always run even if the running of intermediate steps fails.
- If there are multiple steps marked with `defer`, they are run in LIFO order.
    - Also, the included steps are added to run sequence of the parent runbook's deferred steps.
