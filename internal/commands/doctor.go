package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Twigpine/twig/internal/client"
	"github.com/Twigpine/twig/internal/identity"
)

type CheckState int

const (
	StateOk CheckState = iota
	StateWarn
	StateFail
)

type Check struct {
	Label  string
	State  CheckState
	Detail string
	Fix    string
}

// Doctor checks Twigpine installation and connectivity.
func Doctor(nodeURL, dirOverride string) error {
	fmt.Println("twig doctor — checking your Twigpine setup")
	fmt.Println()

	var checks []Check
	allOk := true

	dir, _ := identity.DefaultDir(dirOverride)

	// 1. Identity
	keyPath := identity.KeyPath(dir)
	if kp, err := identity.LoadKeypair(dirOverride); err == nil {
		short := kp.DID()
		if len(short) > 40 {
			short = short[:40] + "…"
		}
		checks = append(checks, Check{
			Label:  "identity",
			State:  StateOk,
			Detail: short,
		})
	} else {
		checks = append(checks, Check{
			Label:  "identity",
			State:  StateFail,
			Detail: fmt.Sprintf("not found at %s", keyPath),
			Fix:    "twig identity new",
		})
	}

	// 2. Registration
	ucanPath := filepath.Join(dir, "ucan.json")
	if data, err := os.ReadFile(ucanPath); err == nil {
		var rec map[string]interface{}
		if json.Unmarshal(data, &rec) == nil {
			node := rec["node"]
			checks = append(checks, Check{
				Label:  "registration",
				State:  StateOk,
				Detail: fmt.Sprintf("registered with %v", node),
			})
		} else {
			checks = append(checks, Check{
				Label:  "registration",
				State:  StateFail,
				Detail: "ucan.json is malformed",
				Fix:    "twig register",
			})
		}
	} else {
		checks = append(checks, Check{
			Label:  "registration",
			State:  StateFail,
			Detail: "not registered with any node",
			Fix:    "twig register",
		})
	}

	// 3. Node env var
	nodeEnv := os.Getenv("TWIGPINE_NODE")
	if nodeEnv == "" {
		nodeEnv = os.Getenv("GITLAWB_NODE")
	}
	if nodeEnv != "" {
		checks = append(checks, Check{
			Label:  "TWIGPINE_NODE",
			State:  StateOk,
			Detail: nodeEnv,
		})
	} else {
		checks = append(checks, Check{
			Label:  "TWIGPINE_NODE",
			State:  StateWarn,
			Detail: "not set — defaulting to " + client.DefaultPublicNode,
			Fix:    "export TWIGPINE_NODE=" + client.DefaultPublicNode,
		})
	}

	// 4. Node connectivity
	nodeURL = client.ResolveNodeURL(nodeURL)
	c := client.New(nodeURL, nil)
	resp, err := c.Get("/")
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var info map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&info)
			v, _ := info["version"].(string)
			d, _ := info["did"].(string)
			if len(d) > 24 {
				d = d[:24] + "…"
			}
			checks = append(checks, Check{
				Label:  "node",
				State:  StateOk,
				Detail: fmt.Sprintf("%s — v%s (%s)", nodeURL, v, d),
			})
		} else {
			checks = append(checks, Check{
				Label:  "node",
				State:  StateFail,
				Detail: fmt.Sprintf("%s returned HTTP %d", nodeURL, resp.StatusCode),
				Fix:    "check TWIGPINE_NODE or target a reachable node",
			})
		}
	} else {
		checks = append(checks, Check{
			Label:  "node",
			State:  StateFail,
			Detail: fmt.Sprintf("%s unreachable: %v", nodeURL, err),
			Fix:    "check your internet connection or node URL",
		})
	}

	// 5. Git remote helper
	helperFound := false
	if _, err := exec.LookPath("git-remote-twigpine"); err == nil {
		helperFound = true
		checks = append(checks, Check{
			Label:  "git-remote-twigpine",
			State:  StateOk,
			Detail: "found in PATH",
		})
	} else if _, err := exec.LookPath("git-remote-gitlawb"); err == nil {
		helperFound = true
		checks = append(checks, Check{
			Label:  "git-remote-gitlawb",
			State:  StateOk,
			Detail: "found in PATH (legacy helper)",
		})
	}

	if !helperFound {
		checks = append(checks, Check{
			Label:  "git-remote-twigpine",
			State:  StateWarn,
			Detail: "not found in PATH — twigpine:// clone/push will use standard git",
			Fix:    "ensure git remote helper is on PATH",
		})
	}

	// 6. Git
	if out, err := exec.Command("git", "--version").Output(); err == nil {
		checks = append(checks, Check{
			Label:  "git",
			State:  StateOk,
			Detail: strings.TrimSpace(string(out)),
		})
	} else {
		checks = append(checks, Check{
			Label:  "git",
			State:  StateFail,
			Detail: "git not found in PATH",
			Fix:    "install git: https://git-scm.com",
		})
	}

	// Render
	for _, ch := range checks {
		icon := "✓"
		switch ch.State {
		case StateWarn:
			icon = "⚠"
		case StateFail:
			icon = "✗"
			allOk = false
		}
		fmt.Printf("  %s  %-24s  %s\n", icon, ch.Label, ch.Detail)
	}

	fmt.Println()
	if allOk {
		fmt.Println("Everything looks good. Run `twig quickstart` to create your first repo.")
	} else {
		fmt.Println("Some checks failed. Suggested fixes:")
		for _, ch := range checks {
			if ch.State == StateFail && ch.Fix != "" {
				fmt.Printf("  %s:  %s\n", ch.Label, ch.Fix)
			}
		}
		fmt.Println("\nRun `twig quickstart` for a guided setup.")
	}

	return nil
}
