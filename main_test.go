package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Twigpine/twig/internal/identity"
)

func TestCLIProcess(t *testing.T) {
	if os.Getenv("TWIG_TEST_CLI_PROCESS") != "1" {
		return
	}
	separator := slices.Index(os.Args, "--")
	if separator < 0 {
		t.Fatal("missing CLI arguments")
	}
	os.Args = append([]string{"twig"}, os.Args[separator+1:]...)
	main()
	os.Exit(0)
}

func runCLI(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	binary := os.Getenv("TWIG_TEST_BINARY")
	if binary == "" {
		binary = os.Args[0]
		args = append([]string{"-test.run=^TestCLIProcess$", "--"}, args...)
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = append(os.Environ(), "TWIG_TEST_CLI_PROCESS=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("CLI timed out: %v", ctx.Err())
	}
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return stdout.String(), stderr.String(), exit.ExitCode()
		}
		t.Fatalf("running CLI: %v", err)
	}
	return stdout.String(), stderr.String(), 0
}

func TestRepoListCLIOutput(t *testing.T) {
	const body = `[{"name":"project","owner_did":"did:key:owner-one","is_public":true,"description":"Example repository","updated_at":"2026-10-04T12:00:00Z","clone_url":"https://example.test/owner-one/project.git","star_count":9007199254740993,"extension":{"enabled":true}}]`
	cases := []struct {
		name string
		args []string
		json bool
	}{
		{name: "default"},
		{name: "table", args: []string{"--format", "table"}},
		{name: "json false", args: []string{"--json=false"}},
		{name: "json", args: []string{"--json"}, json: true},
		{name: "format json", args: []string{"--format", "json"}, json: true},
		{name: "both json", args: []string{"--json", "--format", "json"}, json: true},
		{name: "format json with false shorthand", args: []string{"--json=false", "--format", "json"}, json: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.Method != http.MethodGet || r.URL.RequestURI() != "/api/v1/repos" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.RequestURI())
				}
				if r.Header.Get("Signature") != "" {
					t.Error("anonymous request was signed")
				}
				fmt.Fprint(w, body)
			}))
			defer srv.Close()
			args := append([]string{"repo", "list", "--node", srv.URL, "--dir", t.TempDir()}, tc.args...)
			stdout, stderr, code := runCLI(t, args...)
			if code != 0 || stderr != "" {
				t.Fatalf("exit %d, stderr %q", code, stderr)
			}
			if requests.Load() != 1 {
				t.Fatalf("expected one request, got %d", requests.Load())
			}
			if tc.json {
				var compact bytes.Buffer
				if err := json.Compact(&compact, []byte(stdout)); err != nil {
					t.Fatalf("invalid JSON output: %v", err)
				}
				if compact.String() != body {
					t.Fatalf("JSON data changed: %s", stdout)
				}
				return
			}
			if json.Valid([]byte(stdout)) {
				t.Fatalf("expected a human-readable table, got JSON: %s", stdout)
			}
			for _, value := range []string{"NAME", "OWNER", "VISIBILITY", "UPDATED", "DESCRIPTION", "project", "owner-one", "public", "2026-10-04", "Example repository"} {
				if !strings.Contains(stdout, value) {
					t.Errorf("missing %q from table: %s", value, stdout)
				}
			}
		})
	}
}

func TestRepoListCLIInvalidOptions(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
		code int
	}{
		{name: "unknown format", args: []string{"--format", "yaml"}, want: "unsupported output format", code: 1},
		{name: "empty format", args: []string{"--format="}, want: "unsupported output format", code: 1},
		{name: "conflicting flags", args: []string{"--json", "--format", "table"}, want: "cannot combine", code: 1},
		{name: "reverse conflict", args: []string{"--format", "table", "--json"}, want: "cannot combine", code: 1},
		{name: "unknown format with json", args: []string{"--format", "yaml", "--json"}, want: "unsupported output format", code: 1},
		{name: "help", args: []string{"--help"}, want: "-format", code: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				fmt.Fprint(w, "[]")
			}))
			defer srv.Close()
			args := append([]string{"repo", "list", "--node", srv.URL, "--dir", t.TempDir()}, tc.args...)
			stdout, stderr, code := runCLI(t, args...)
			if code != tc.code || stdout != "" || !strings.Contains(stderr, tc.want) {
				t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
			}
			if tc.name == "help" && !strings.Contains(stderr, "-json") {
				t.Errorf("JSON flag missing from help: %s", stderr)
			}
			if requests.Load() != 0 {
				t.Errorf("invalid flags or help contacted the node")
			}
		})
	}
}

func TestRepoListCLIErrors(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{name: "message", status: 403, body: `{"message":"denied\u001b‮","error":"other"}`, want: "request failed (403): denied"},
		{name: "error", status: 404, body: `{"error":"missing"}`, want: "request failed (404): missing"},
		{name: "plain text", status: 502, body: "unavailable", want: "request failed (502): unavailable"},
		{name: "malformed", status: 200, body: `[`, want: "decoding repositories"},
		{name: "object", status: 200, body: `{}`, want: "decoding repositories"},
		{name: "null", status: 200, body: `null`, want: "expected an array"},
		{name: "invalid row", status: 200, body: `[null]`, want: "expected an object"},
	}
	for _, format := range []string{"table", "json"} {
		for _, tc := range cases {
			t.Run(format+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.status)
					fmt.Fprint(w, tc.body)
				}))
				defer srv.Close()
				stdout, stderr, code := runCLI(t, "repo", "list", "--node", srv.URL, "--dir", t.TempDir(), "--format", format)
				if code != 1 || stdout != "" || !strings.Contains(stderr, tc.want) {
					t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
				}
				if strings.ContainsAny(stderr, "\x1b‮") {
					t.Errorf("unsafe error output: %q", stderr)
				}
			})
		}
	}
}

func TestRepoListCLISignedRequest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kp, err := identity.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	if err := identity.SaveKeypair(dir, kp); err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != http.MethodGet || r.URL.RequestURI() != "/api/v1/repos" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.RequestURI())
		}
		if r.Header.Get("Signature") == "" || !strings.Contains(r.Header.Get("Signature-Input"), kp.DID()) {
			t.Error("request missing identity signature")
		}
		fmt.Fprint(w, "[]")
	}))
	defer srv.Close()
	stdout, stderr, code := runCLI(t, "repo", "list", "--node", srv.URL, "--dir", dir)
	if code != 0 || stderr != "" || stdout != "No repositories found.\n" || requests.Load() != 1 {
		t.Fatalf("exit %d, requests %d, stdout %q, stderr %q", code, requests.Load(), stdout, stderr)
	}
}
