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

#### Run N runbooks

``` go
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http/httptest"
	"testing"

	"github.com/k1LoW/runn"
	_ "github.com/lib/pq"
)

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

#### Run single runbook

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


## Runbook

The runbook file has the following format.

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
            Authorization: "token {{ steps[1].res.session_token }}"
          body: nil
  -
    test: steps[3].res.status == 200
  -
    test: len(steps[3].res.projects) > 0
```

#### Grouping of related parts by color

![color](docs/runbook.svg)

> Documentation is WIP

## Alternatives

- [zoncoen/scenarigo](https://github.com/zoncoen/scenarigo): An end-to-end scenario testing tool for HTTP/gRPC server.
