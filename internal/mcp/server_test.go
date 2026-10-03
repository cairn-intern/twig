package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

func TestReadWriteMessageRoundTrip(t *testing.T) {
	orig := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "ping",
	}

	var buf bytes.Buffer
	if err := WriteMessage(&buf, orig); err != nil {
		t.Fatalf("WriteMessage error: %v", err)
	}

	reader := bufio.NewReader(&buf)
	msgBytes, err := ReadMessage(reader)
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}

	var recovered map[string]interface{}
	if err := json.Unmarshal(msgBytes, &recovered); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	if recovered["method"] != "ping" || recovered["id"] != float64(1) {
		t.Fatalf("recovered msg mismatch: %v", recovered)
	}
}

func TestServerInitialize(t *testing.T) {
	srv := NewServer("https://node.twigpine.com", nil)

	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	go func() {
		_ = srv.Serve(inR, outW)
		outW.Close()
	}()

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
	}

	go func() {
		_ = WriteMessage(inW, req)
		inW.Close()
	}()

	reader := bufio.NewReader(outR)
	respBytes, err := ReadMessage(reader)
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			ServerInfo struct {
				Name string `json:"name"`
			} `json:"serverInfo"`
		} `json:"result"`
	}

	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.Result.ServerInfo.Name != "twig" {
		t.Fatalf("expected server name 'twig', got %s", resp.Result.ServerInfo.Name)
	}
}
