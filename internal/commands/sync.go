package commands

import (
	"fmt"

	"github.com/Twigpine/twig/internal/client"
)

// SyncTrigger pulls repos from known peers into the sync queue.
func SyncTrigger(nodeURL, dirOverride string) error {
	kp, err := EnsureIdentityExists(dirOverride)
	if err != nil {
		return err
	}

	nodeURL = client.ResolveNodeURL(nodeURL)
	c := client.New(nodeURL, kp)

	resp, err := c.Post("/api/v1/sync/trigger", []byte("{}"))
	if err != nil {
		return fmt.Errorf("sync trigger failed: %w", err)
	}

	return PrintResponseOrError(resp)
}

// SyncStatus shows the current sync queue status.
func SyncStatus(nodeURL string) error {
	nodeURL = client.ResolveNodeURL(nodeURL)
	c := client.New(nodeURL, nil)

	resp, err := c.Get("/api/v1/sync/status")
	if err != nil {
		return fmt.Errorf("sync status failed: %w", err)
	}

	return PrintResponseOrError(resp)
}
