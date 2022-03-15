# runn

`runn` ( means "Run N" ) is a package/tool for running operations following a scenario.

Key features of `runn` are:

- **As a tool for scenario testing.**
- **As a test helper package for the Go language.**
- **As a tool for automation.**

## Usage

`runn` can run a multi-step scenario following a `runbook` written in YAML format.

### As a tool for scenario testing / As a tool for automation.

`runn` can run one or more runbooks.

``` console
$ runn list path/to/**/*.yml
  Desc                     Path
-----------------------------------------------------
  Login and get projects.  path/to/books/login.yml
  Login and logout.        path/to/books/logout.yml
  New project.             path/to/books/new.yml
$ runn run path/to/**/*.yml
Login and get projects. ... ok
Login and logout. ... ok
New project. ... ok

3 scenarios, 0 failures
```

### As a test helper package for the Go language.

`runn` can also behave as a test helper for the Go language.

#### Run N runbooks using httptest.Server

``` go
func TestRouter(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("postgres", "user=root password=root host=localhost dbname=test sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewServer(NewRouter(db))
	t.Cleanup(func() {
		ts.Close()
		db.Close()
	})
	o, err := runn.Load("testdata/books/**/*.yml", runn.T(t), runn.Runner("req", ts.URL), runn.DBRunner("db", db))
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
	db, err := sql.Open("postgres", "user=root password=root host=localhost dbname=test sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewServer(NewRouter(db))
	t.Cleanup(func() {
		ts.Close()
		db.Close()
	})
	o, err := runn.New(runn.T(t), runn.Book("testdata/books/login.yml"), runn.Runner("req", ts.URL), runn.DBRunner("db", db))
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
	db, err := sql.Open("postgres", "user=root password=root host=localhost dbname=test sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	o, err := runn.Load("testdata/books/**/*.yml", runn.T(t), runn.HTTPRunnerWithHandler("req", NewRouter(db)), runn.DBRunner("db", db))
	if err != nil {
		t.Fatal(err)
	}
	if err := o.RunN(ctx); err != nil {
		t.Fatal(err)
	}
}
```

## Runbook

The runbook file has the following format.

`step:` section accepts array or ordered map.

### Array

``` yaml
desc: Login and get projects.
runners:
  req: https://example.com/api/v1
  db: mysql://root:mypass@localhost:3306/testdb
vars:
  username: alice
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
              password: "{{ steps[0].rows[0].password }}"
  -
    test: steps[1].res.status == 200
  -
    req:
      /projects:
        get:
          headers:
            Authorization: "token {{ steps[1].res.body.session_token }}"
          body: null
  -
    test: steps[3].res.status == 200
  -
    test: len(steps[3].res.body.projects) > 0
```

### Map

``` yaml
desc: Login and get projects.
runners:
  req: https://example.com/api/v1
  db: mysql://root:mypass@localhost:3306/testdb
vars:
  username: alice
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
              password: "{{ steps.find_user.rows[0].password }}"
  test_status0:
    test: steps.login.res.status == 200
  list_projects:
    req:
      /projects:
        get:
          headers:
            Authorization: "token {{ steps.login.res.body.session_token }}"
          body: null
  test_status1:
    test: steps.list_projects.res.status == 200
  count_projects:
    test: len(steps.list_projects.res.body.projects) > 0
```

#### Grouping of related parts by color

![color](docs/runbook.svg)

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

## Runner

### HTTP Runner: Do HTTP request

Use `https://` or `http://` scheme to specify HTTP Runner.

When the step is invoked, it sends the specified HTTP Request and records the response.

``` yaml
runners:
  ghapi: https://api.github.com
```

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

### Dump Runner: dump recorded values

The `dump` runner is a built-in runner, so there is no need to specify it in the `runners:` section.

It dumps the specified recorded values.

``` yaml
-
  dump: steps[4].rows
```

### Include Runner: include other runbook

The `include` runner is a built-in runner, so there is no need to specify it in the `runners:` section.

Include runner reads and runs the runbook in the specified path.

Recorded values are nested.

``` yaml
-
  include: path/to/get_token.yml
```

## Install

### As tool

**deb:**

Use [dpkg-i-from-url](https://github.com/k1LoW/dpkg-i-from-url)

``` console
$ export RUNN_VERSION=X.X.X
$ curl -L https://git.io/dpkg-i-from-url | bash -s -- https://github.com/k1LoW/runn/releases/download/v$RUNN_VERSION/runn_$RUNN_VERSION-1_amd64.deb
```

**RPM:**

``` console
$ export RUNN_VERSION=X.X.X
$ yum install https://github.com/k1LoW/runn/releases/download/v$RUNN_VERSION/runn_$RUNN_VERSION-1_amd64.rpm
```

**apk:**

Use [apk-add-from-url](https://github.com/k1LoW/apk-add-from-url)

``` console
$ export RUNN_VERSION=X.X.X
$ curl -L https://git.io/apk-add-from-url | sh -s -- https://github.com/k1LoW/runn/releases/download/v$RUNN_VERSION/runn_$RUNN_VERSION-1_amd64.apk
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

### As test helper

```console
$ go get github.com/k1LoW/runn
```

## Alternatives

- [zoncoen/scenarigo](https://github.com/zoncoen/scenarigo): An end-to-end scenario testing tool for HTTP/gRPC server.
