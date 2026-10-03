package commands

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Twigpine/twig/internal/client"
	"github.com/Twigpine/twig/internal/did"
	"github.com/Twigpine/twig/internal/identity"
)

// CertList lists ref certificates for a repository.
func CertList(repoInput, nodeURL, dirOverride string) error {
	owner, name, err := ResolveRepoOwner(repoInput, dirOverride)
	if err != nil {
		return err
	}

	kp, _ := identity.LoadKeypair(dirOverride)
	nodeURL = client.ResolveNodeURL(nodeURL)
	c := client.New(nodeURL, kp)

	resp, err := c.GetAuthed(fmt.Sprintf("/api/v1/repos/%s/%s/certs", owner, name))
	if err != nil {
		return fmt.Errorf("listing certs: %w", err)
	}

	return PrintResponseOrError(resp)
}

// CertShow retrieves a certificate and optionally verifies its signature.
func CertShow(repoInput, certID, nodeURL, dirOverride string, verify bool, expectNode string) error {
	owner, name, err := ResolveRepoOwner(repoInput, dirOverride)
	if err != nil {
		return err
	}

	kp, _ := identity.LoadKeypair(dirOverride)
	nodeURL = client.ResolveNodeURL(nodeURL)
	c := client.New(nodeURL, kp)

	resp, err := c.GetAuthed(fmt.Sprintf("/api/v1/repos/%s/%s/certs/%s", owner, name, certID))
	if err != nil {
		return fmt.Errorf("fetching cert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fetching cert failed (%d): %s", resp.StatusCode, client.SanitizeNodeMsg(string(data)))
	}

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var certObj map[string]interface{}
	if err := json.Unmarshal(rawBytes, &certObj); err != nil {
		return err
	}

	if verify {
		nodeDID, _ := certObj["node_did"].(string)
		if nodeDID == "" {
			return errors.New("certificate missing node_did field")
		}

		if expectNode != "" && nodeDID != expectNode {
			return fmt.Errorf("certificate node_did %s does not match expected %s", nodeDID, expectNode)
		}

		pubKey, err := did.ToVerifyingKey(nodeDID)
		if err != nil {
			return fmt.Errorf("invalid issuer node DID: %w", err)
		}

		sigB64, _ := certObj["signature"].(string)
		if sigB64 == "" {
			return errors.New("certificate missing signature field")
		}

		sigBytes, err := base64.StdEncoding.DecodeString(sigB64)
		if err != nil {
			sigBytes, err = base64.RawURLEncoding.DecodeString(sigB64)
			if err != nil {
				return fmt.Errorf("invalid signature base64: %w", err)
			}
		}

		// Re-marshal signed payload fields
		payload := map[string]interface{}{}
		for k, v := range certObj {
			if k != "signature" {
				payload[k] = v
			}
		}
		payloadBytes, _ := json.Marshal(payload)

		if !identity.Verify(pubKey, payloadBytes, sigBytes) {
			return errors.New("certificate signature is INVALID")
		}

		fmt.Println("✓ Certificate signature is VALID")
	}

	pretty, _ := json.MarshalIndent(certObj, "", "  ")
	fmt.Println(string(pretty))
	return nil
}
