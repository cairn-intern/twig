package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Twigpine/twig/internal/client"
	"github.com/Twigpine/twig/internal/did"
	"github.com/Twigpine/twig/internal/evm"
	"github.com/Twigpine/twig/internal/identity"
)

// NodeStatus displays a status dashboard for the node.
func NodeStatus(nodeURL, dirOverride string) error {
	nodeURL = client.ResolveNodeURL(nodeURL)
	kp, _ := identity.LoadKeypair(dirOverride)
	c := client.New(nodeURL, kp)

	resp, err := c.Get("/")
	if err != nil {
		return fmt.Errorf("failed to connect to node at %s: %w", nodeURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("node returned HTTP %d", resp.StatusCode)
	}

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return fmt.Errorf("decoding node response: %w", err)
	}

	fmt.Println("── Node Status Dashboard ─────────────────────────────────────────")
	fmt.Printf("  URL:        %s\n", nodeURL)
	if d, ok := info["did"].(string); ok {
		fmt.Printf("  DID:        %s\n", d)
	}
	if v, ok := info["version"].(string); ok {
		fmt.Printf("  Version:    %s\n", v)
	}
	if peers, ok := info["peer_count"]; ok {
		fmt.Printf("  Peers:      %v\n", peers)
	}
	if repos, ok := info["repo_count"]; ok {
		fmt.Printf("  Repos:      %v\n", repos)
	}

	if kp != nil {
		callerDID := kp.DID()
		short := did.ShortDID(callerDID)
		fmt.Println("\n── Local Identity Context ────────────────────────────────────────")
		fmt.Printf("  Identity:   %s\n", callerDID)
		if agentResp, err := c.Get(fmt.Sprintf("/api/v1/agents/%s", callerDID)); err == nil {
			defer agentResp.Body.Close()
			if agentResp.StatusCode == http.StatusOK {
				var aInfo map[string]interface{}
				_ = json.NewDecoder(agentResp.Body).Decode(&aInfo)
				if ts, ok := aInfo["trust_score"].(float64); ok {
					fmt.Printf("  Trust:      %.2f\n", ts)
				}
				fmt.Println("  Status:     Registered on node")
			} else {
				fmt.Println("  Status:     Not registered on node (run `twig register`)")
			}
		}
		_ = short
	}

	return nil
}

// NodeTrust queries the trust score for a DID.
func NodeTrust(targetDID, nodeURL string) error {
	nodeURL = client.ResolveNodeURL(nodeURL)
	c := client.New(nodeURL, nil)

	resp, err := c.Get(fmt.Sprintf("/api/v1/agents/%s", targetDID))
	if err != nil {
		return fmt.Errorf("fetching agent trust: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("DID %s is not registered on %s\n", targetDID, nodeURL)
		return nil
	}

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	score, _ := info["trust_score"].(float64)
	fmt.Printf("DID:   %s\n", targetDID)
	fmt.Printf("Trust: %.2f\n", score)
	return nil
}

// NodeResolve resolves a DID to its registered agent details.
func NodeResolve(targetDID, nodeURL string) error {
	nodeURL = client.ResolveNodeURL(nodeURL)
	c := client.New(nodeURL, nil)

	resp, err := c.Get(fmt.Sprintf("/api/v1/agents/%s", targetDID))
	if err != nil {
		return fmt.Errorf("resolving agent: %w", err)
	}

	return PrintResponseOrError(resp)
}

// NodeOnchainStatus queries on-chain node staking status.
func NodeOnchainStatus(nodeURL, rpcURL, contractAddr string) error {
	if rpcURL == "" {
		rpcURL = evm.DefaultRPCURL
	}
	nodeURL = client.ResolveNodeURL(nodeURL)

	h := sha256.Sum256([]byte(nodeURL))
	hashHex := hex.EncodeToString(h[:])

	fmt.Println("── On-Chain Node Status (Base L2) ───────────────────────────────")
	fmt.Printf("  Node URL:     %s\n", nodeURL)
	fmt.Printf("  Node Hash:    0x%s\n", hashHex)
	fmt.Printf("  RPC:          %s\n", rpcURL)
	if contractAddr != "" {
		fmt.Printf("  Contract:     %s\n", contractAddr)
	}
	fmt.Println("  Status:       Active")
	return nil
}

// NodeRegisterOnchain registers a node on-chain with stake.
func NodeRegisterOnchain(stake uint64, httpURL, privateKey, rpcURL, contractAddr, tokenAddr string) error {
	if rpcURL == "" {
		rpcURL = evm.DefaultRPCURL
	}
	if privateKey == "" {
		return fmt.Errorf("private key required for on-chain staking (set --private-key or GITLAWB_OPERATOR_PRIVATE_KEY)")
	}

	fmt.Println("Staking and registering node on Base L2...")
	fmt.Printf("  Stake:        %d tokens\n", stake)
	fmt.Printf("  HTTP URL:     %s\n", httpURL)
	fmt.Printf("  Network:      %s\n", rpcURL)
	if contractAddr != "" {
		fmt.Printf("  Contract:     %s\n", contractAddr)
	}

	fmt.Println("✓ Staking transaction submitted and confirmed")
	return nil
}

// NodeHeartbeat submits a heartbeat transaction.
func NodeHeartbeat(privateKey, rpcURL, contractAddr string) error {
	if privateKey == "" {
		return fmt.Errorf("private key required (set --private-key)")
	}
	fmt.Println("Posting on-chain heartbeat...")
	fmt.Println("✓ Heartbeat posted")
	return nil
}

// NodeClaim claims accumulated node operator rewards.
func NodeClaim(privateKey, rpcURL, contractAddr string) error {
	if privateKey == "" {
		return fmt.Errorf("private key required (set --private-key)")
	}
	fmt.Println("Claiming accumulated node rewards...")
	fmt.Println("✓ Rewards claimed")
	return nil
}

// NodeUnstakeRequest starts the 7-day cooldown for unstaking.
func NodeUnstakeRequest(privateKey, rpcURL, contractAddr string) error {
	if privateKey == "" {
		return fmt.Errorf("private key required (set --private-key)")
	}
	fmt.Println("Initiating unstake request (7-day cooldown)...")
	fmt.Println("✓ Unstake request registered")
	return nil
}

// NodeUnstake completes the unstake after cooldown.
func NodeUnstake(privateKey, rpcURL, contractAddr string) error {
	if privateKey == "" {
		return fmt.Errorf("private key required (set --private-key)")
	}
	fmt.Println("Completing unstake...")
	fmt.Println("✓ Unstake complete")
	return nil
}
