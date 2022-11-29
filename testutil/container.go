package testutil

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/k1LoW/sshc/v3"
	"github.com/ory/dockertest/v3"
	"golang.org/x/crypto/ssh"
)

func CreateHTTPBinContainer(t *testing.T) string {
	t.Helper()
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}
	httpbin, err := pool.Run("kennethreitz/httpbin", "latest", []string{})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}
	t.Cleanup(func() {
		if err := pool.Purge(httpbin); err != nil {
			t.Fatalf("Could not purge resource: %s", err)
		}
	})

	var host string
	if err := pool.Retry(func() error {
		host = fmt.Sprintf("http://localhost:%s/", httpbin.GetPort("80/tcp"))
		_, err := http.Get(host)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("Could not connect to database: %s", err)
	}

	return host
}

func CreateMySQLContainer(t *testing.T) *sql.DB {
	t.Helper()
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	opt := &dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "8",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=rootpass",
			"MYSQL_USER=myuser",
			"MYSQL_PASSWORD=mypass",
			"MYSQL_DATABASE=testdb",
		},
		Mounts: []string{
			fmt.Sprintf("%s:/docker-entrypoint-initdb.d/initdb.sql", filepath.Join(wd, "testdata", "initdb", "mysql.sql")),
		},
		Cmd: []string{
			"mysqld",
			"--character-set-server=utf8mb4",
			"--collation-server=utf8mb4_unicode_ci",
		},
	}
	my, err := pool.RunWithOptions(opt)
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}
	t.Cleanup(func() {
		if err := pool.Purge(my); err != nil {
			t.Fatalf("Could not purge resource: %s", err)
		}
	})

	var db *sql.DB
	if err := pool.Retry(func() error {
		time.Sleep(time.Second * 10)
		var err error
		db, err = sql.Open("mysql", fmt.Sprintf("myuser:mypass@(localhost:%s)/testdb?parseTime=true", my.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		t.Fatalf("Could not connect to database: %s", err)
	}

	return db
}

func CreateSSHdContainer(t *testing.T) (*ssh.Client, string, string, string, int) {
	t.Helper()
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	opt := &dockertest.RunOptions{
		Repository: "panubo/sshd",
		Tag:        "latest",
		Env: []string{
			"SSH_USERS=testuser:1000:1000",
		},
		Mounts: []string{
			fmt.Sprintf("%s:/keys", filepath.Join(wd, "testdata", "sshd")),
			fmt.Sprintf("%s:/etc/entrypoint.d", filepath.Join(wd, "testdata", "sshd", "entrypoint.d")),
		},
	}
	sshd, err := pool.RunWithOptions(opt)
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}
	t.Cleanup(func() {
		if err := pool.Purge(sshd); err != nil {
			t.Fatalf("Could not purge resource: %s", err)
		}
	})

	var (
		client *ssh.Client
		port   int
	)
	host := "myserver"
	if err := pool.Retry(func() error {
		port, err = strconv.Atoi(sshd.GetPort("22/tcp"))
		if err != nil {
			return err
		}
		client, err = sshc.NewClient(host, sshc.ConfigPath(filepath.Join(wd, "testdata", "sshd", "ssh_config")), sshc.Port(port))
		if err != nil {
			return err
		}
		sess, err := client.NewSession()
		if err != nil {
			return err
		}
		defer sess.Close()
		if err := sess.Run("pwd"); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("Could not connect to sshd: %s", err)
	}

	return client, host, "127.0.0.1", "testuser", port
}
