<p align="center">
<img src="https://github.com/k1LoW/runn/raw/main/docs/logo.svg" width="200" alt="runn">
</p>

`runn` ( means "Run N" ) is a package/tool for running operations following a scenario.

Key features of `runn` are:

- **As a tool for scenario based testing.**
- **As a test helper package for the Go language.**
- **As a tool for automation.**
- **OpenAPI Document-like syntax for HTTP request testing.**

## Usage

`runn` can run a multi-step scenario following a `runbook` written in YAML format.

### As a tool for scenario based testing / As a tool for automation.

`runn` can run one or more runbooks as a CLI tool.

``` console
$ runn list path/to/**/*.yml
  Desc                               Path                               If
---------------------------------------------------------------------------------
  Login and get projects.            pato/to/book/projects.yml
  Login and logout.                  pato/to/book/logout.yml
  Only if included.                  pato/to/book/only_if_included.yml  included
$ runn run path/to/**/*.yml
Login and get projects. ... ok
Login and logout. ... ok
Only if included. ... skip

3 scenarios, 1 skipped, 0 failures
```

### As a test helper package for the Go language.

`runn` can also behave as a test helper for the Go language.

#### Run N runbooks using httptest.Server

``` go
func TestRouter(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/testdb")
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewServer(NewRouter(db))
	t.Cleanup(func() {
		ts.Close()
		db.Close()
	})
	opts := []runn.Option{
		runn.T(t),
		runn.Runner("req", ts.URL),
		runn.DBRunner("db", db),
	}
	o, err := runn.Load("testdata/books/**/*.yml", opts...)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.RunN(ctx); err != nil {
		t.Fatal(err)
	}
}
```

#### Run single runbook using httptest.Server

``` go
func TestRouter(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/testdb")
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewServer(NewRouter(db))
	t.Cleanup(func() {
		ts.Close()
		db.Close()
	})
	opts := []runn.Option{
		runn.T(t), 
		runn.Book("testdata/books/login.yml"),
		runn.Runner("req", ts.URL),
		runn.DBRunner("db", db),
	}
	o, err := runn.New(opts...)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Run(ctx); err != nil {
		t.Fatal(err)
	}
}
```

#### Run N runbooks with http.Handler

``` go
func TestRouter(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/testdb")
	if err != nil {
		log.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	opts := []runn.Option{
		runn.T(t),
		runn.HTTPRunnerWithHandler("req", NewRouter(db)),
		runn.DBRunner("db", db),
	}
	o, err := runn.Load("testdata/books/**/*.yml", opts...)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.RunN(ctx); err != nil {
		t.Fatal(err)
	}
}
```

## Runbook ( runn scenario file )

The runbook file has the following format.

`step:` section accepts **array** or **ordered map**.

**Array:**

``` yaml
desc: Login and get projects.
runners:
  req: https://example.com/api/v1
  db: mysql://root:mypass@localhost:3306/testdb
vars:
  username: alice
  password: ${TEST_PASS}
steps:
  -
    db:
      query: SELECT * FROM users WHERE name = '{{ vars.username }}'
  -
    req:
      /login:
        post:
          body:
            application/json:
              email: "{{ steps[0].rows[0].email }}"
              password: "{{ vars.password }}"
    test: steps[1].res.status == 200
  -
    req:
      /projects:
        get:
          headers:
            Authorization: "token {{ steps[1].res.body.session_token }}"
          body: null
    test: steps[2].res.status == 200
  -
    test: len(steps[2].res.body.projects) > 0
```

**Map:**

``` yaml
desc: Login and get projects.
runners:
  req: https://example.com/api/v1
  db: mysql://root:mypass@localhost:3306/testdb
vars:
  username: alice
  password: ${TEST_PASS}
steps:
  find_user:
    db:
      query: SELECT * FROM users WHERE name = '{{ vars.username }}'
  login:
    req:
      /login:
        post:
          body:
            application/json:
              email: "{{ steps.find_user.rows[0].email }}"
              password: "{{ vars.password }}"
    test: steps.login.res.status == 200
  list_projects:
    req:
      /projects:
        get:
          headers:
            Authorization: "token {{ steps.login.res.body.session_token }}"
          body: null
    test: steps.list_projects.res.status == 200
  count_projects:
    test: len(steps.list_projects.res.body.projects) > 0
```

#### Grouping of related parts by color

**Array:**

![color](docs/runbook.svg)

**Map:**

![color](docs/runbook_map.svg)

### `desc:`

Description of runbook.

### `runners:`

Mapping of runners that run `steps:` of runbook.

In the `steps:` section, call the runner with the key specified in the `runners:` section.

Built-in runners such as test runner do not need to be specified in this section.

``` yaml
runners:
  ghapi: ${GITHUB_API_ENDPOINT}
  idp: https://auth.example.com
  db: my:dbuser:${DB_PASS}@hostname:3306/dbname
```

In the example, each runner can be called by `ghapi:`, `idp:` or `db:` in `steps:`.

### `vars:`

Mapping of variables available in the `steps:` of runbook.

``` yaml
vars:
  username: alice@example.com
  token: ${SECRET_TOKEN}
```

In the example, each variable can be used in `{{ vars.username }}` or `{{ vars.token }}` in `steps:`.

### `debug:`

Enable debug output for runn.

``` yaml
debug: true
```

### `if:`

Conditions for skip all steps.

``` yaml
if: included # Run steps only if included
```

### `skipTest:`

Skip all `test:` sections

``` yaml
skipTest: true
```

### `steps:`

Steps to run in runbook.

The steps are invoked in order from top to bottom.

Any return values are recorded for each step.

When `steps:` is array, recorded values can be retrieved with `{{ steps[*].* }}`.

``` yaml
steps:
  -
    db:
      query: SELECT * FROM users WHERE name = '{{ vars.username }}'
  -
    req:
      /users/{{ steps[0].rows[0].id }}:
        get:
          body: null
```

When `steps:` is map, recorded values can be retrieved with `{{ steps.<key>.* }}`.

``` yaml
steps:
  find_user:
    db:
      query: SELECT * FROM users WHERE name = '{{ vars.username }}'
  user_info:
    req:
      /users/{{ steps.find_user.rows[0].id }}:
        get:
          body: null
```

### `steps[*].desc:` `steps.<key>.desc:`

Description of step.

### `steps[*].if:` `steps.<key>.if:`

Conditions for skip step.

``` yaml
steps:
  login:
    if: 'len(vars.token) == 0' # Run step only if var.token is not set
    req:
      /login:
        post:
          body:
[...]
```

### `steps[*].retry:` `steps.<key>.retry:`

Retry settings for steps.

``` yaml
steps:
  waitingroom:
    retry:
      count: 10
      until: 'steps.waitingroom.res.status == "201"'
      minInterval: 0.5 # sec
      maxInterval: 10  # sec
      # jitter: 0.0
      # interval: 5
      # multiplier: 1.5
    req:
      /cart/in:
        post:
          body:
[...]
```

## Runner

### HTTP Runner: Do HTTP request

Use `https://` or `http://` scheme to specify HTTP Runner.

When the step is invoked, it sends the specified HTTP Request and records the response.

``` yaml
runners:
  ghapi: https://api.github.com
```

#### Validation of HTTP request and HTTP response

HTTP requests sent by `runn` and their HTTP responses can be validated.

**OpenAPI v3:**

``` yaml
runners:
  myapi:
    endpoint: https://api.github.com
    openapi3: path/to/openapi.yaml
    # skipValidateRequest: false
    # skipValidateResponse: false
```

### gRPC Runner: Do gRPC request

Use `grpc://` scheme to specify gRPC Runner.

When the step is invoked, it sends the specified gRPC Request and records the response.

``` yaml
runners:
  greq: grpc://grpc.example.com:80
```

See [testdata/book/grpc.yml](testdata/book/grpc.yml).

### DB Runner: Query a database

Use dsn (Data Source Name) to specify DB Runner.

When step is executed, it executes the specified query the database.

If the query is a SELECT clause, it records the selected `rows`, otherwise it records `last_insert_id` and `rows_affected` .

#### Support Databases

**PostgreSQL:**

``` yaml
runners:
  mydb: postgres://dbuser:dbpass@hostname:5432/dbname
```

``` yaml
runners:
  db: pg://dbuser:dbpass@hostname:5432/dbname
```

**MySQL:**

``` yaml
runners:
  testdb: mysql://dbuser:dbpass@hostname:3306/dbname
```

``` yaml
runners:
  db: my://dbuser:dbpass@hostname:3306/dbname
```

**SQLite3:**

``` yaml
runners:
  db: sqlite:///path/to/dbname.db
```

``` yaml
runners:
  local: sq://dbname.db
```

### Exec Runner: execute command

The `exec` runner is a built-in runner, so there is no need to specify it in the `runners:` section.

It execute command using `command:` and `stdin:`

``` yaml
-
  exec:
    command: grep error
    stdin: '{{ steps[3].res.rawBody }}'
```

### Test Runner: test using recorded values

The `test` runner is a built-in runner, so there is no need to specify it in the `runners:` section.

It evaluates the conditional expression using the recorded values.

``` yaml
-
  test: steps[3].res.status == 200
```

The `test` runner can run in the same steps as the other runners.

### Dump Runner: dump recorded values

The `dump` runner is a built-in runner, so there is no need to specify it in the `runners:` section.

It dumps the specified recorded values.

``` yaml
-
  dump: steps[4].rows
```

The `dump` runner can run in the same steps as the other runners.

### Include Runner: include other runbook

The `include` runner is a built-in runner, so there is no need to specify it in the `runners:` section.

Include runner reads and runs the runbook in the specified path.

Recorded values are nested.

``` yaml
-
  include: path/to/get_token.yml
```

It is also possible to override `vars:` of included runbook.

``` yaml
-
  include:
    path: path/to/login.yml
    vars:
      username: alice
      password: alicepass
-
  include:
    path: path/to/login.yml
    vars:
      username: bob
      password: bobpass
```

It is also possible to skip all `test:` sections in the included runbook.

``` yaml
-
  include:
    path: path/to/signup.yml
    skipTest: true
```

### Bind Runner: bind variables

The `bind` runner is a built-in runner, so there is no need to specify it in the `runners:` section.

It bind runner binds any values with another key.

``` yaml
  -
    req:
      /users/k1low:
        get:
          body: null
  -
    bind:
      user_id: steps[0].res.body.data.id
  -
    dump: user_id
```

The `bind` runner can run in the same steps as the other runners.

## Option

See https://pkg.go.dev/github.com/k1LoW/runn#Option

### Example: Run as a test helper ( func `T` )

https://pkg.go.dev/github.com/k1LoW/runn#T

``` go
o, err := runn.Load("testdata/**/*.yml", runn.T(t))
if err != nil {
	t.Fatal(err)
}
if err := o.RunN(ctx); err != nil {
	t.Fatal(err)
}
```

### Example: Add custom function ( func `Func` )

https://pkg.go.dev/github.com/k1LoW/runn#Func

``` yaml
desc: Test using GitHub
runners:
  req:
    endpoint: https://github.com
steps:
  -
    req:
      /search?l={{ urlencode('C++') }}&q=runn&type=Repositories:
        get:
          body:
            application/json:
              null
    test: 'steps[0].res.status == 200'
```

``` go
o, err := runn.Load("testdata/**/*.yml", runn.Func("urlencode", url.QueryEscape))
if err != nil {
	t.Fatal(err)
}
if err := o.RunN(ctx); err != nil {
	t.Fatal(err)
}
```

## Filter runbooks to be executed by the environment variable `RUNN_RUN`

Run only runbooks matching the filename "login".

``` console
$ env RUNN_RUN=login go test ./... -run TestRouter
```

## Install

### As a CLI tool

**deb:**

``` console
$ export RUNN_VERSION=X.X.X
$ curl -o runn.deb -L https://github.com/k1LoW/runn/releases/download/v$RUNN_VERSION/runn_$RUNN_VERSION-1_amd64.deb
$ dpkg -i runn.deb
```

**RPM:**

``` console
$ export RUNN_VERSION=X.X.X
$ yum install https://github.com/k1LoW/runn/releases/download/v$RUNN_VERSION/runn_$RUNN_VERSION-1_amd64.rpm
```

**apk:**

``` console
$ export RUNN_VERSION=X.X.X
$ curl -o runn.apk -L https://github.com/k1LoW/runn/releases/download/v$RUNN_VERSION/runn_$RUNN_VERSION-1_amd64.apk
$ apk add runn.apk
```

**homebrew tap:**

```console
$ brew install k1LoW/tap/runn
```

**manually:**

Download binary from [releases page](https://github.com/k1LoW/runn/releases)

**docker:**

```console
$ docker pull ghcr.io/k1low/runn:latest
```

**go install:**

```console
$ go install github.com/k1LoW/runn/cmd/runn@latest
```

### As a test helper

```console
$ go get github.com/k1LoW/runn
```

## Alternatives

- [zoncoen/scenarigo](https://github.com/zoncoen/scenarigo): An end-to-end scenario testing tool for HTTP/gRPC server.

## References

- [zoncoen/scenarigo](https://github.com/zoncoen/scenarigo): An end-to-end scenario testing tool for HTTP/gRPC server.
