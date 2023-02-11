<p align="center">
<img src="https://github.com/k1LoW/runn/raw/main/docs/logo.svg" width="200" alt="runn">
</p>

[![build](https://github.com/k1LoW/runn/actions/workflows/ci.yml/badge.svg)](https://github.com/k1LoW/runn/actions/workflows/ci.yml) ![Coverage](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/runn/coverage.svg) ![Code to Test Ratio](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/runn/ratio.svg) ![Test Execution Time](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/runn/time.svg)

`runn` ( means "Run N". is pronounced /rʌ́n én/. ) is a package/tool for running operations following a scenario.

Key features of `runn` are:

- **As a tool for scenario based testing.**
- **As a test helper package for the Go language.**
- **As a tool for workflow automation.**
- **Support HTTP request, gRPC request, DB query, Chrome DevTools Protocol, and SSH/Local command execution**
- **OpenAPI Document-like syntax for HTTP request testing.**
- **Single binary = CI-Friendly.**

## Online book

- [runn cookbook (Japanese)](https://zenn.dev/k1low/books/runn-cookbook)

## Quickstart

You can use the `runn new` command to quickly start creating scenarios ([runbooks](#runbook--runn-scenario-file-)).

**:rocket: Create and run scenario using `curl` or `grpcurl` commands:**

![docs/runn.svg](docs/runn.svg)

<details>

<summary>Command details</summary>

``` console
$ curl https://httpbin.org/json -H "accept: application/json"
{
  "slideshow": {
    "author": "Yours Truly",
    "date": "date of publication",
    "slides": [
      {
        "title": "Wake up to WonderWidgets!",
        "type": "all"
      },
      {
        "items": [
          "Why <em>WonderWidgets</em> are great",
          "Who <em>buys</em> WonderWidgets"
        ],
        "title": "Overview",
        "type": "all"
      }
    ],
    "title": "Sample Slide Show"
  }
}
$ runn new --and-run --desc 'httpbin.org GET' --out http.yml -- curl https://httpbin.org/json -H "accept: application/json"
$ grpcurl -d '{"greeting": "alice"}' grpcb.in:9001 hello.HelloService/SayHello
{
  "reply": "hello alice"
}
$ runn new --and-run --desc 'grpcb.in Call' --out grpc.yml -- grpcurl -d '{"greeting": "alice"}' grpcb.in:9001 hello.HelloService/SayHello
$ runn list *.yml
  Desc             Path      If
---------------------------------
  grpcb.in Call    grpc.yml
  httpbin.org GET  http.yml
$ runn run *.yml
grpcb.in Call ... ok
httpbin.org GET ... ok

2 scenarios, 0 skipped, 0 failures
```

</details>

**:rocket: Create scenario using access log:**

![docs/runn_axslog.svg](docs/runn_axslog.svg)

<details>

<summary>Command details</summary>

``` console
$ cat access_log
183.87.255.54 - - [18/May/2019:05:37:09 +0200] "GET /?post=%3script%3ealert(1); HTTP/1.0" 200 42433
62.109.16.162 - - [18/May/2019:05:37:12 +0200] "GET /core/files/js/editor.js/?form=\xeb\x2a\x5e\x89\x76\x08\xc6\x46\x07\x00\xc7\x46\x0c\x00\x00\x00\x80\xe8\xdc\xff\xff\xff/bin/sh HTTP/1.0" 200 81956
87.251.81.179 - - [18/May/2019:05:37:13 +0200] "GET /login.php/?user=admin&amount=100000 HTTP/1.0" 400 4797
103.36.79.144 - - [18/May/2019:05:37:14 +0200] "GET /authorize.php/.well-known/assetlinks.json HTTP/1.0" 200 9436
$ cat access_log| runn new --out axslog.yml
$ cat axslog.yml| yq
desc: Generated by `runn new`
runners:
  req: https://dummy.example.com
steps:
  - req:
      /?post=%3script%3ealert(1);:
        get:
          body: null
  - req:
      /core/files/js/editor.js/?form=xebx2ax5ex89x76x08xc6x46x07x00xc7x46x0cx00x00x00x80xe8xdcxffxffxff/bin/sh:
        get:
          body: null
  - req:
      /login.php/?user=admin&amount=100000:
        get:
          body: null
  - req:
      /authorize.php/.well-known/assetlinks.json:
        get:
          body: null
$
```

</details>

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

#### Run N runbooks using [httptest.Server](https://pkg.go.dev/net/http/httptest#Server) and [sql.DB](https://pkg.go.dev/database/sql#DB)

``` go
func TestRouter(t *testing.T) {
	ctx := context.Background()
	dsn := "username:password@tcp(localhost:3306)/testdb"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	dbr, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewServer(NewRouter(db))
	t.Cleanup(func() {
		ts.Close()
		db.Close()
		dbr.Close()
	})
	opts := []runn.Option{
		runn.T(t),
		runn.Runner("req", ts.URL),
		runn.DBRunner("db", dbr),
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

#### Run single runbook using [httptest.Server](https://pkg.go.dev/net/http/httptest#Server) and [sql.DB](https://pkg.go.dev/database/sql#DB)

``` go
func TestRouter(t *testing.T) {
	ctx := context.Background()
	dsn := "username:password@tcp(localhost:3306)/testdb"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	dbr, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewServer(NewRouter(db))
	t.Cleanup(func() {
		ts.Close()
		db.Close()
		dbr.Close()
	})
	opts := []runn.Option{
		runn.T(t),
		runn.Book("testdata/books/login.yml"),
		runn.Runner("req", ts.URL),
		runn.DBRunner("db", dbr),
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

#### Run N runbooks using [grpc.Server](https://pkg.go.dev/google.golang.org/grpc#Server)

``` go
func TestServer(t *testing.T) {
	addr := "127.0.0.1:8080"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	ts := grpc.NewServer()
	myapppb.RegisterMyappServiceServer(s, NewMyappServer())
	reflection.Register(s)
	go func() {
		s.Serve(l)
	}()
	t.Cleanup(func() {
		ts.GracefulStop()
	})
	opts := []runn.Option{
		runn.T(t),
		runn.Runner("greq", fmt.Sprintf("grpc://%s", addr),
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

#### Run N runbooks with [http.Handler](https://pkg.go.dev/net/http#Handler) and [sql.DB](https://pkg.go.dev/database/sql#DB)

``` go
func TestRouter(t *testing.T) {
	ctx := context.Background()
	dsn := "username:password@tcp(localhost:3306)/testdb"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	dbr, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		dbr.Close()
	})
	opts := []runn.Option{
		runn.T(t),
		runn.HTTPRunnerWithHandler("req", NewRouter(db)),
		runn.DBRunner("db", dbr),
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

## Examples

See the [details](./examples)

## Runbook ( runn scenario file )

The runbook file has the following format.

`step:` section accepts **list** or **ordered map**.

**List:**

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

**List:**

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

### `loop:`

Loop setting for runbook.

#### Simple loop runbook

``` yaml
loop: 10
steps:
  [...]
```

or

``` yaml
loop:
  count: 10
steps:
  [...]
```

#### Retry runbook

It can be used as a retry mechanism by setting a condition in the `until:` section.

If the condition of `until:` is met, the loop is broken without waiting for the number of `count:` to be run.

Also, if the run of the number of `count:` completes but does not satisfy the condition of `until:`, then the step is considered to be failed.

``` yaml
loop:
  count: 10
  until: 'outcome == "success"' # until the runbook outcome is successful.
  minInterval: 0.5 # sec
  maxInterval: 10  # sec
  # jitter: 0.0
  # interval: 5
  # multiplier: 1.5
steps:
  waitingroom:
    req:
      /cart/in:
        post:
          body:
[...]
```

- `outcome` ... the result of a completed (`success`, `failure`, `skipped`).

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

### `steps[*].loop:` `steps.<key>.loop:`

Loop settings for steps.

#### Simple loop step

``` yaml
steps:
  multicartin:
    loop: 10
    req:
      /cart/in:
        post:
          body:
            application/json:
              product_id: "{{ i }}" # The loop count (0..9) is assigned to `i`.
[...]
```

or

``` yaml
steps:
  multicartin:
    loop:
      count: 10
    req:
      /cart/in:
        post:
          body:
            application/json:
              product_id: "{{ i }}" # The loop count (0..9) is assigned to `i`.
[...]
```

#### Retry step

It can be used as a retry mechanism by setting a condition in the `until:` section.

If the condition of `until:` is met, the loop is broken without waiting for the number of `count:` to be run.

Also, if the run of the number of `count:` completes but does not satisfy the condition of `until:`, then the step is considered to be failed.

``` yaml
steps:
  waitingroom:
    loop:
      count: 10
      until: 'steps.waitingroom.res.status == "201"' # Store values of latest loop
      minInterval: 500ms
      maxInterval: 10 # sec
      # jitter: 0.0
      # interval: 5
      # multiplier: 1.5
    req:
      /cart/in:
        post:
          body:
[...]
```

( `steps[*].retry:` `steps.<key>.retry:` are deprecated )

## Runner

### HTTP Runner: Do HTTP request

Use `https://` or `http://` scheme to specify HTTP Runner.

When the step is invoked, it sends the specified HTTP Request and records the response.

``` yaml
runners:
  req: https://example.com
steps:
  -
    desc: Post /users                     # description of step
    req:                                  # key to identify the runner. In this case, it is HTTP Runner.
      /users:                             # path of http request
        post:                             # method of http request
          headers:                        # headers of http request
            Authorization: 'Bearer xxxxx'
          body:                           # body of http request
            application/json:             # Content-Type specification. In this case, it is "Content-Type: application/json"
              username: alice
              password: passw0rd
    test: |                               # test for current step
      current.res.status == 201
```

See [testdata/book/http.yml](testdata/book/http.yml) and [testdata/book/http_multipart.yml](testdata/book/http_multipart.yml).

#### Structure of recorded responses

The following response

```
HTTP/1.1 200 OK
Content-Length: 29
Content-Type: application/json
Date: Wed, 07 Sep 2022 06:28:20 GMT

{"data":{"username":"alice"}}
```

is recorded with the following structure.

``` yaml
[`step key` or `current` or `previous`]:
  res:
    status: 200                              # current.res.status
    headers:
      Content-Length:
        - '29'                               # current.res.headers["Content-Length"][0]
      Content-Type:
        - 'application/json'                 # current.res.headers["Content-Type"][0]w
      Date:
        - 'Wed, 07 Sep 2022 06:28:20 GMT'    # current.res.headers["Date"][0]
    body:
      data:
        username: 'alice'                    # current.res.body.data.username
    rawBody: '{"data":{"username":"alice"}}' # current.res.rawBody
```

#### Do not follow redirect

The HTTP Runner interprets HTTP responses and automatically redirects.
To disable this, set `notFollowRedirect` to true.

``` yaml
runners:
  req:
    endpoint: https://example.com
    notFollowRedirect: true
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

#### Custom CA and Certificates

``` yaml
runners:
  myapi:
    endpoint: https://api.github.com
    cacert: path/to/cacert.pem
    cert: path/to/cert.pem
    key: path/to/key.pem
```

### gRPC Runner: Do gRPC request

Use `grpc://` scheme to specify gRPC Runner.

When the step is invoked, it sends the specified gRPC Request and records the response.

``` yaml
runners:
  greq: grpc://grpc.example.com:80
steps:
  -
    desc: Request using Unary RPC                     # description of step
    greq:                                             # key to identify the runner. In this case, it is gRPC Runner.
      grpctest.GrpcTestService/Hello:                 # package.Service/Method of rpc
        headers:                                      # headers of rpc
          authentication: tokenhello
        message:                                      # message of rpc
          name: alice
          num: 3
          request_time: 2022-06-25T05:24:43.861872Z
  -
    desc: Request using Client streaming RPC
    greq:
      grpctest.GrpcTestService/MultiHello:
        headers:
          authentication: tokenmultihello
        messages:                                     # messages of rpc
          -
            name: alice
            num: 5
            request_time: 2022-06-25T05:24:43.861872Z
          -
            name: bob
            num: 6
            request_time: 2022-06-25T05:24:43.861872Z
```

``` yaml
runners:
  greq:
    addr: grpc.example.com:8080
    tls: true
    cacert: path/to/cacert.pem
    cert: path/to/cert.pem
    key: path/to/key.pem
    # skipVerify: false
```

See [testdata/book/grpc.yml](testdata/book/grpc.yml).

#### Structure of recorded responses

The following response

```protocol-buffer
message HelloResponse {
  string message = 1;

  int32 num = 2;

  google.protobuf.Timestamp create_time = 3;
}
```

```json
{"create_time":"2022-06-25T05:24:43.861872Z","message":"hello","num":32}
```


and headers

```yaml
content-type: ["application/grpc"]
hello: ["this is header"]
```

and trailers

```yaml
hello: ["this is trailer"]
```

are recorded with the following structure.

``` yaml
[`step key` or `current` or `previous`]:
  res:
    status: 0                                      # current.res.status
    headers:
      content-type:
        - 'application/grpc'                       # current.res.headers[0].content-type
      hello:
        - 'this is header'                         # current.res.headers[0].hello
    trailers:
      hello:
        - 'this is trailer'                        # current.res.trailers[0].hello
    message:
      create_time: '2022-06-25T05:24:43.861872Z'   # current.res.message.create_time
      message: 'hello'                             # current.res.message.message
      num: 32                                      # current.res.message.num
    messages:
      -
        create_time: '2022-06-25T05:24:43.861872Z' # current.res.messages[0].create_time
        message: 'hello'                           # current.res.messages[0].message
        num: 32                                    # current.res.messages[0].num
```

### DB Runner: Query a database

Use dsn (Data Source Name) to specify DB Runner.

When step is invoked, it executes the specified query the database.

``` yaml
runners:
  db: postgres://dbuser:dbpass@hostname:5432/dbname
steps:
  -
    desc: Select users            # description of step
    db:                           # key to identify the runner. In this case, it is DB Runner.
      query: SELECT * FROM users; # query to execute
```

See [testdata/book/db.yml](testdata/book/db.yml).

#### Structure of recorded responses

If the query is a SELECT clause, it records the selected `rows`,

``` yaml
[`step key` or `current` or `previous`]:
  rows:
    -
      id: 1                           # current.rows[0].id
      username: 'alice'               # current.rows[0].username
      password: 'passw0rd'            # current.rows[0].password
      email: 'alice@example.com'      # current.rows[0].email
      created: '2017-12-05T00:00:00Z' # current.rows[0].created
    -
      id: 2                           # current.rows[1].id
      username: 'bob'                 # current.rows[1].username
      password: 'passw0rd'            # current.rows[1].password
      email: 'bob@example.com'        # current.rows[1].email
      created: '2022-02-22T00:00:00Z' # current.rows[1].created
```

otherwise it records `last_insert_id` and `rows_affected` .

``` yaml
[`step key` or `current` or `previous`]:
  last_insert_id: 3 # current.last_insert_id
  rows_affected: 1  # current.rows_affected
```

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

### CDP Runner: Control browser using Chrome DevTools Protocol (CDP)

Use `cdp://` or `chrome://` scheme to specify CDP Runner.

When the step is invoked, it controls browser via Chrome DevTools Protocol.

``` yaml
runners:
  cc: chrome://new
steps:
  -
    desc: Navigate, click and get h1 using CDP  # description of step
    cc:                                         # key to identify the runner. In this case, it is CDP Runner.
      actions:                                  # actions to control browser
        - navigate: https://pkg.go.dev/time
        - click: 'body > header > div.go-Header-inner > nav > div > ul > li:nth-child(2) > a'
        - waitVisible: 'body > footer'
        - text: 'h1'
  -
    test: |
      previous.text == 'Install the latest version of Go'
```

See [testdata/book/cdp.yml](testdata/book/cdp.yml).

#### Functions for action to control browser

<!-- repin:fndoc -->
**`attributes`** (aliases: `getAttributes`, `attrs`, `getAttrs`)

Get the element attributes for the first element node matching the selector (`sel`).

```yaml
actions:
  - attributes:
      sel: 'h1'
# record to current.attrs:
```

or

```yaml
actions:
  - attributes: 'h1'
```

**`click`**

Send a mouse click event to the first element node matching the selector (`sel`).

```yaml
actions:
  - click:
      sel: 'nav > div > a'
```

or

```yaml
actions:
  - click: 'nav > div > a'
```

**`doubleClick`**

Send a mouse double click event to the first element node matching the selector (`sel`).

```yaml
actions:
  - doubleClick:
      sel: 'nav > div > li'
```

or

```yaml
actions:
  - doubleClick: 'nav > div > li'
```

**`evaluate`** (aliases: `eval`)

Evaluate the Javascript expression (`expr`).

```yaml
actions:
  - evaluate:
      expr: 'document.querySelector("h1").textContent = "hello"'
```

or

```yaml
actions:
  - evaluate: 'document.querySelector("h1").textContent = "hello"'
```

**`fullHTML`** (aliases: `getFullHTML`, `getHTML`, `html`)

Get the full html of page.

```yaml
actions:
  - fullHTML
# record to current.html:
```

**`innerHTML`** (aliases: `getInnerHTML`)

Get the inner html of the first element node matching the selector (`sel`).

```yaml
actions:
  - innerHTML:
      sel: 'h1'
# record to current.html:
```

or

```yaml
actions:
  - innerHTML: 'h1'
```

**`latestTab`** (aliases: `latestTarget`)

Change current frame to latest tab.

```yaml
actions:
  - latestTab
```

**`localStorage`** (aliases: `getLocalStorage`)

Get localStorage items.

```yaml
actions:
  - localStorage:
      origin: 'https://github.com'
# record to current.items:
```

or

```yaml
actions:
  - localStorage: 'https://github.com'
```

**`location`** (aliases: `getLocation`)

Get the document location.

```yaml
actions:
  - location
# record to current.url:
```

**`navigate`**

Navigate the current frame to `url` page.

```yaml
actions:
  - navigate:
      url: 'https://pkg.go.dev/time'
```

or

```yaml
actions:
  - navigate: 'https://pkg.go.dev/time'
```

**`outerHTML`** (aliases: `getOuterHTML`)

Get the outer html of the first element node matching the selector (`sel`).

```yaml
actions:
  - outerHTML:
      sel: 'h1'
# record to current.html:
```

or

```yaml
actions:
  - outerHTML: 'h1'
```

**`screenshot`** (aliases: `getScreenshot`)

Take a full screenshot of the entire browser viewport.

```yaml
actions:
  - screenshot
# record to current.png:
```

**`scroll`** (aliases: `scrollIntoView`)

Scroll the window to the first element node matching the selector (`sel`).

```yaml
actions:
  - scroll:
      sel: 'body > footer'
```

or

```yaml
actions:
  - scroll: 'body > footer'
```

**`sendKeys`**

Send keys (`value`) to the first element node matching the selector (`sel`).

```yaml
actions:
  - sendKeys:
      sel: 'input[name=username]'
      value: 'k1lowxb@gmail.com'
```

**`sessionStorage`** (aliases: `getSessionStorage`)

Get sessionStorage items.

```yaml
actions:
  - sessionStorage:
      origin: 'https://github.com'
# record to current.items:
```

or

```yaml
actions:
  - sessionStorage: 'https://github.com'
```

**`setUploadFile`** (aliases: `setUpload`)

Set upload file (`path`) to the first element node matching the selector (`sel`).

```yaml
actions:
  - setUploadFile:
      sel: 'input[name=avator]'
      path: '/path/to/image.png'
```

**`setUserAgent`** (aliases: `setUA`, `ua`, `userAgent`)

Set the default User-Agent

```yaml
actions:
  - setUserAgent:
      userAgent: 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36'
```

or

```yaml
actions:
  - setUserAgent: 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36'
```

**`submit`**

Submit the parent form of the first element node matching the selector (`sel`).

```yaml
actions:
  - submit:
      sel: 'form.login'
```

or

```yaml
actions:
  - submit: 'form.login'
```

**`text`** (aliases: `getText`)

Get the visible text of the first element node matching the selector (`sel`).

```yaml
actions:
  - text:
      sel: 'h1'
# record to current.text:
```

or

```yaml
actions:
  - text: 'h1'
```

**`textContent`** (aliases: `getTextContent`)

Get the text content of the first element node matching the selector (`sel`).

```yaml
actions:
  - textContent:
      sel: 'h1'
# record to current.text:
```

or

```yaml
actions:
  - textContent: 'h1'
```

**`title`** (aliases: `getTitle`)

Get the document `title`.

```yaml
actions:
  - title
# record to current.title:
```

**`value`** (aliases: `getValue`)

Get the Javascript value field of the first element node matching the selector (`sel`).

```yaml
actions:
  - value:
      sel: 'input[name=address]'
# record to current.value:
```

or

```yaml
actions:
  - value: 'input[name=address]'
```

**`wait`** (aliases: `sleep`)

Wait for the specified `time`.

```yaml
actions:
  - wait:
      time: '10sec'
```

or

```yaml
actions:
  - wait: '10sec'
```

**`waitReady`**

Wait until the element matching the selector (`sel`) is ready.

```yaml
actions:
  - waitReady:
      sel: 'body > footer'
```

or

```yaml
actions:
  - waitReady: 'body > footer'
```

**`waitVisible`**

Wait until the element matching the selector (`sel`) is visible.

```yaml
actions:
  - waitVisible:
      sel: 'body > footer'
```

or

```yaml
actions:
  - waitVisible: 'body > footer'
```


<!-- repin:fndoc -->

### SSH Runner: execute commands on a remote server connected via SSH

Use `ssh://` scheme to specify SSH Runner.

When step is invoked, it executes commands on a remote server connected via SSH.

``` yaml
runners:
  sc: ssh://username@hostname:port
steps:
  -
    desc: 'execute `hostname`' # description of step
    sc:
      command: hostname
```


``` yaml
runners:
  sc:
    hostname: hostname
    user: username
    port: 22
    # host: myserver
    # sshConfig: path/to/ssh_config
    # keepSession: false
    # localForward: '33306:127.0.0.1:3306'
    # keyboardInteractive:
    #   - match: Username
    #     answer: k1low
    #   - match: OTP
    #     answer: ${MY_OTP}
```

See [testdata/book/sshd.yml](testdata/book/sshd.yml).

#### Structure of recorded responses

The response to the run command is always `stdout` and `stderr`.

``` yaml
[`step key` or `current` or `previous`]:
  stdout: 'hello world' # current.stdout
  stderr: ''            # current.stderr
```

### Exec Runner: execute command

The `exec` runner is a built-in runner, so there is no need to specify it in the `runners:` section.

It execute command using `command:` and `stdin:`

``` yaml
-
  exec:
    command: grep hello
    stdin: '{{ steps[3].res.rawBody }}'
```

See [testdata/book/exec.yml](testdata/book/exec.yml).

#### Structure of recorded responses

The response to the run command is always `stdout`, `stderr` and `exit_code`.

``` yaml
[`step key` or `current` or `previous`]:
  stdout: 'hello world' # current.stdout
  stderr: ''            # current.stderr
  exit_code: 0          # current.exit_code
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

or

``` yaml
-
  dump:
    expr: steps[4].rows
    out: path/to/dump.out
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

## Expression evaluation engine

runn has embedded [antonmedv/expr](https://github.com/antonmedv/expr) as the evaluation engine for the expression.

See [Language Definition](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md).

### Built-in functions

- `urlencode` ... [url.QueryEscape](https://pkg.go.dev/net/url#QueryEscape)
- `base64encode` ... [base64.EncodeToString](https://pkg.go.dev/encoding/base64#Encoding.EncodeToString)
- `base64decode` ... [base64.DecodeString](https://pkg.go.dev/encoding/base64#Encoding.DecodeString)
- `string` ... [cast.ToString](https://pkg.go.dev/github.com/spf13/cast#ToString)
- `int` ... [cast.ToInt](https://pkg.go.dev/github.com/spf13/cast#ToInt)
- `bool` ... [cast.ToBool](https://pkg.go.dev/github.com/spf13/cast#ToBool)
- `compare` ... Compare two values ( `func(x, y interface{}, ignoreKeys ...string) bool` ).
- `diff` ... Difference between two values ( `func(x, y interface{}, ignoreKeys ...string) string` ).
- `input` ... [prompter.Prompt](https://pkg.go.dev/github.com/Songmu/prompter#Prompt)
- `intersect` ... Find the intersection of two iterable values ( `func(x, y interface{}) interface{}` ).
- `secret` ... [prompter.Password](https://pkg.go.dev/github.com/Songmu/prompter#Password)
- `select` ... [prompter.Choose](https://pkg.go.dev/github.com/Songmu/prompter#Choose)
- `basename` ... [filepath.Base](https://pkg.go.dev/path/filepath#Base)

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

## Measure elapsed time as profile

``` go
opts := []runn.Option{
	runn.T(t),
	runn.Book("testdata/books/login.yml"),
	runn.Profile(true)
}
o, err := runn.New(opts...)
if err != nil {
	t.Fatal(err)
}
if err := o.Run(ctx); err != nil {
	t.Fatal(err)
}
f, err := os.Open("profile.json")
if err != nil {
	t.Fatal(err)
}
if err := o.DumpProfile(f); err != nil {
	t.Fatal(err)
}
```

or

``` console
$ runn run testdata/books/login.yml --profile
```

The runbook run profile can be read with `runn rprof` command.

``` console
$ runn rprof runn.prof
  runbook[login site](t/b/login.yml)           2995.72ms
    steps[0].req                                747.67ms
    steps[1].req                                185.69ms
    steps[2].req                                192.65ms
    steps[3].req                                188.23ms
    steps[4].req                                569.53ms
    steps[5].req                                299.88ms
    steps[6].test                                 0.14ms
    steps[7].include                            620.88ms
      runbook[include](t/b/login_include.yml)   605.56ms
        steps[0].req                            605.54ms
    steps[8].req                                190.92ms
  [total]                                      2995.84ms
```

## Capture runbook runs

``` go
opts := []runn.Option{
	runn.T(t),
	runn.Capture(capture.Runbook("path/to/dir")),
}
o, err := runn.Load("testdata/books/**/*.yml", opts...)
if err != nil {
	t.Fatal(err)
}
if err := o.RunN(ctx); err != nil {
	t.Fatal(err)
}
```

or

``` console
$ runn run path/to/**/*.yml --capture path/to/dir
```

## Load test using runbooks

You can use the `runn loadt` command for load testing using runbooks.

``` console
$ runn loadt --concurrent 2 path/to/*.yml

Number of runbooks per RunN...: 15
Warm up time (--warm-up)......: 5s
Duration (--duration).........: 10s
Concurrent (--concurrent).....: 2

Total.........................: 12
Succeeded.....................: 12
Failed........................: 0
Error rate....................: 0%
RunN per seconds..............: 1.2
Latency ......................: max=1,835.1ms min=1,451.3ms avg=1,627.8ms med=1,619.8ms p(90)=1,741.5ms p(99)=1,788.4ms

```

It also checks the results of the load test with the `--threshold` option. If the condition is not met, it returns exit status 1.

``` console
$ runn loadt --concurrent 2 --threshold 'error_rate < 10' path/to/*.yml

Number of runbooks per RunN...: 15
Warm up time (--warm-up)......: 5s
Duration (--duration).........: 10s
Concurrent (--concurrent).....: 2

Total.........................: 13
Succeeded.....................: 12
Failed........................: 1
Error rate....................: 7.6%
RunN per seconds..............: 1.3
Latency ......................: max=1,790.2ms min=95.0ms avg=1,541.4ms med=1,640.4ms p(90)=1,749.7ms p(99)=1,786.5ms

Error: (error_rate < 10) is not true
error_rate < 10
├── error_rate => 14.285714285714285
└── 10 => 10
```

### Variables for threshold

| Variable name | Type | Description |
| --- | --- | --- |
| `total` | `int` | Total |
| `succeeded` | `int` | Succeeded |
| `failed` | `int` | Failed |
| `error_rate` | `float` | Error rate |
| `rps` | `float` | RunN per seconds |
| `max` | `float` | Latency max (ms) |
| `mid` | `float` | Latency mid (ms) |
| `min` | `float` | Latency min (ms) |
| `p90` | `float` | Latency p(90) (ms) |
| `p99` | `float` | Latency p(99) (ms) |
| `avg` | `float` | Latency avg (ms) |

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
$ docker container run -it --rm --name runn -v $PWD:/books ghcr.io/k1low/runn:latest list /books/*.yml
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
- [fullstorydev/grpcurl](https://github.com/fullstorydev/grpcurl): Like cURL, but for gRPC: Command-line tool for interacting with gRPC servers
