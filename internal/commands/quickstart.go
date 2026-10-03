package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Twigpine/twig/internal/client"
	"github.com/Twigpine/twig/internal/identity"
)

// Quickstart runs an interactive setup wizard.
func Quickstart(nodeURL, dirOverride string, skipPrompts bool) error {
	fmt.Println()
	fmt.Println("Welcome to Twigpine.")
	fmt.Println("This wizard will set up your identity, register you with a node,")
	fmt.Println("and create your first repository.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Step 1: Identity
	fmt.Println("── Step 1: Identity ─────────────────────────────────────────────────")
	fmt.Println()
	kp, err := identity.LoadKeypair(dirOverride)
	if err == nil {
		fmt.Println("  ✓ Identity already exists")
		fmt.Printf("    DID: %s\n\n", kp.DID())
	} else {
		if !skipPrompts {
			fmt.Print("Generate a new identity keypair? [Y/n] ")
			ans, _ := reader.ReadString('\n')
			ans = strings.TrimSpace(strings.ToLower(ans))
			if ans == "n" || ans == "no" {
				fmt.Println("Quickstart aborted.")
				return nil
			}
		}
		if err := IdentityNew(dirOverride, false, os.Stdin); err != nil {
			return err
		}
	}

	// Step 2: Register
	nodeURL = client.ResolveNodeURL(nodeURL)
	fmt.Println("── Step 2: Registration ─────────────────────────────────────────────")
	fmt.Println()
	if err := Register(nodeURL, nil, "", dirOverride); err != nil {
		fmt.Printf("Registration notice: %v\n", err)
	}

	// Step 3: First repo
	fmt.Println()
	fmt.Println("── Step 3: First Repository ─────────────────────────────────────────")
	fmt.Println()
	repoName := "my-first-repo"
	if !skipPrompts {
		fmt.Printf("Repository name [%s]: ", repoName)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			repoName = input
		}
	}

	fmt.Printf("Creating %s...\n", repoName)
	if err := RepoCreate(repoName, "Created with Twigpine quickstart", false, "main", nodeURL, dirOverride); err != nil {
		fmt.Printf("Notice creating repo: %v\n", err)
	} else {
		fmt.Printf("✓ Created %s\n", repoName)
	}

	fmt.Println()
	fmt.Println("── Next Steps ───────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("You're ready to use Twigpine!")
	fmt.Println("Clone your new repo:")
	fmt.Printf("  twig repo clone %s\n\n", repoName)
	fmt.Println("Or check system status:")
	fmt.Println("  twig status")

	return nil
}
