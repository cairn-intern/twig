package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/Twigpine/twig/internal/client"
	"github.com/Twigpine/twig/internal/did"
)

type repoListRow struct {
	Name        string `json:"name"`
	OwnerDID    string `json:"owner_did"`
	IsPublic    *bool  `json:"is_public"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updated_at"`
}

func writeRepoList(w io.Writer, data []byte, format string) error {
	if format != "table" && format != "json" {
		return fmt.Errorf("unsupported output format %q (expected table or json)", format)
	}
	var repos []json.RawMessage
	if err := json.Unmarshal(data, &repos); err != nil {
		return fmt.Errorf("decoding repositories: %w", err)
	}
	if repos == nil {
		return fmt.Errorf("decoding repositories: expected an array")
	}
	for i, repo := range repos {
		if bytes.TrimSpace(repo)[0] != '{' {
			return fmt.Errorf("decoding repository %d: expected an object", i+1)
		}
	}
	if format == "json" {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, bytes.TrimSpace(data), "", "  "); err != nil {
			return fmt.Errorf("formatting repositories: %w", err)
		}
		_, err := fmt.Fprintln(w, pretty.String())
		return err
	}
	rows := make([]repoListRow, len(repos))
	for i, repo := range repos {
		if err := json.Unmarshal(repo, &rows[i]); err != nil {
			return fmt.Errorf("decoding repository %d: %w", i+1, err)
		}
	}
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "No repositories found.")
		return err
	}

	table := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(table, "NAME\tOWNER\tVISIBILITY\tUPDATED\tDESCRIPTION"); err != nil {
		return err
	}
	for _, repo := range rows {
		name := client.SanitizeNodeMsg(repo.Name)
		if name == "" {
			name = "-"
		}
		owner := client.SanitizeNodeMsg(did.ShortDID(repo.OwnerDID))
		if owner == "" {
			owner = "-"
		}
		visibility := "-"
		if repo.IsPublic != nil {
			visibility = "private"
			if *repo.IsPublic {
				visibility = "public"
			}
		}
		updated := "-"
		if len(repo.UpdatedAt) >= len(time.DateOnly) {
			if date, err := time.Parse(time.DateOnly, repo.UpdatedAt[:len(time.DateOnly)]); err == nil {
				updated = date.Format(time.DateOnly)
			}
		}
		if _, err := fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\n", name, owner, visibility, updated, client.SanitizeNodeMsg(repo.Description)); err != nil {
			return err
		}
	}
	return table.Flush()
}
