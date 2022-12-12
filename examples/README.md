# Runnable Examples

- [HTTP](#http-runner)
- [Expression and Built-in Func](#expression-evaluation-and-built-in-function)
- [Chrome](#chrome)
- [Include](#include)
- [Go Test](#go-test-helper)

## HTTP Runner

[http.yml](./http.yml)

```
$ runn run ./http.yml --debug
```

## Expression evaluation and Built-in function

For the [details](../README.md#expression-evaluation-engine)

[expr.yml](./expr.yml)

```
// expression evaluation engine
$ runn run ./expr.yml --debug
```

[func.yml](./func.yml)

```
// built-in function
$ runn run ./func.yml --debug
```

## Chrome

[cdp.yml](./cdp.yml)

```
$ runn run ./cdp.yml --debug
```

## Include

[include.yml](./include.yml)

```
$ runn run ./include.yml --debug
```

## Go Test Helper

[go-test](./go-test)

```
$ cd go-test
$ go test -cover
```
