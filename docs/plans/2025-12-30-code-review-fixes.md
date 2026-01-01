# Code Review Fixes Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix all important and minor issues identified in the code review.

**Architecture:** Small, focused fixes to existing files. No new packages or major refactoring.

**Tech Stack:** Go 1.25, cobra CLI

---

## Task 1: Fix Division by Zero in Quota Display

**Files:**
- Modify: `internal/cmd/message.go:220-223`

**Step 1: Fix the division**

In `getMessageQuota`, add a guard for zero quota value:

```go
// Replace lines 220-223 in internal/cmd/message.go
if quota.Type == "limited" && quota.Value > 0 {
    pct := float64(consumption.TotalUsage) / float64(quota.Value) * 100
    fmt.Fprintf(cmd.OutOrStdout(), "Message Quota: %d/month\n", quota.Value)
    fmt.Fprintf(cmd.OutOrStdout(), "Used: %d (%.1f%%)\n", consumption.TotalUsage, pct)
} else if quota.Type == "limited" {
    fmt.Fprintf(cmd.OutOrStdout(), "Message Quota: 0/month\n")
    fmt.Fprintf(cmd.OutOrStdout(), "Used: %d\n", consumption.TotalUsage)
} else {
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/cmd/message.go
git commit -m "fix: prevent division by zero in quota display"
```

---

## Task 2: Add Token Validation in Auth Flow

**Files:**
- Modify: `internal/auth/server.go:103-116`

**Step 1: Add validation before storing**

Add empty token check after parsing form values:

```go
// After line 104 in internal/auth/server.go, add:
if accessToken == "" {
    http.Error(w, "Access token is required", http.StatusBadRequest)
    return
}
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/auth/server.go
git commit -m "fix: validate access token is not empty in auth flow"
```

---

## Task 3: Add Broadcast Confirmation

**Files:**
- Modify: `internal/cmd/message.go:66-91`

**Step 1: Add confirmation prompt**

Update `newMessageBroadcastCmd` to check for `--yes` flag and prompt if not set:

```go
// Replace the RunE function in newMessageBroadcastCmd (lines 79-91)
RunE: func(cmd *cobra.Command, args []string) error {
    if text == "" && flexJSON == "" {
        return fmt.Errorf("specify --text or --flex")
    }
    if text != "" && flexJSON != "" {
        return fmt.Errorf("specify either --text or --flex, not both")
    }

    // Require confirmation for broadcast unless --yes is set
    if !flags.Yes {
        fmt.Fprint(cmd.OutOrStdout(), "This will broadcast to ALL followers. Continue? [y/N]: ")
        var response string
        fmt.Fscanln(cmd.InOrStdin(), &response)
        if response != "y" && response != "Y" && response != "yes" {
            return fmt.Errorf("broadcast cancelled")
        }
    }

    if text != "" {
        return broadcastTextMessage(cmd, text)
    }
    return broadcastFlexMessage(cmd, flexJSON)
},
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/cmd/message.go
git commit -m "feat: add confirmation prompt for broadcast command"
```

---

## Task 4: Add Rich Menu Size Option

**Files:**
- Modify: `internal/cmd/richmenu.go:41-66`

**Step 1: Add size flag and update create function**

Add a `--size` flag to the create command:

```go
// Replace newRichMenuCreateCmd function
func newRichMenuCreateCmd() *cobra.Command {
    var chatBarText string
    var actionsJSON string
    var size string

    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new rich menu",
        Long:  "Create a rich menu with the specified actions and chat bar text.",
        Example: `  # Create a full-size rich menu
  line richmenu create --name "Main Menu" --actions '[{"type":"message","label":"Help","text":"help"}]'

  # Create a compact rich menu
  line richmenu create --name "Menu" --size compact --actions '[...]'`,
        RunE: func(cmd *cobra.Command, args []string) error {
            if chatBarText == "" {
                return fmt.Errorf("--name is required")
            }
            if actionsJSON == "" {
                return fmt.Errorf("--actions is required")
            }
            if size != "full" && size != "compact" {
                return fmt.Errorf("--size must be 'full' or 'compact'")
            }
            return createRichMenu(cmd, chatBarText, actionsJSON, size)
        },
    }

    cmd.Flags().StringVar(&chatBarText, "name", "", "Chat bar text / menu name (required)")
    cmd.Flags().StringVar(&actionsJSON, "actions", "", "Actions JSON array (required)")
    cmd.Flags().StringVar(&size, "size", "full", "Menu size: full (2500x1686) or compact (2500x843)")

    return cmd
}
```

**Step 2: Update createRichMenu function signature and implementation**

Find and update the `createRichMenu` function to accept size parameter:

```go
// Update createRichMenu function signature and size handling
func createRichMenu(cmd *cobra.Command, name, actionsJSON, size string) error {
    client, err := newAPIClient()
    if err != nil {
        return err
    }

    var actions []json.RawMessage
    if err := json.Unmarshal([]byte(actionsJSON), &actions); err != nil {
        return fmt.Errorf("invalid actions JSON: %w", err)
    }

    // Determine dimensions based on size
    height := 1686
    if size == "compact" {
        height = 843
    }

    areaWidth := 2500 / len(actions)
    areas := make([]api.RichMenuArea, len(actions))
    for i, action := range actions {
        areas[i] = api.RichMenuArea{
            Bounds: api.RichMenuBounds{
                X:      i * areaWidth,
                Y:      0,
                Width:  areaWidth,
                Height: height,
            },
            Action: action,
        }
    }

    req := api.CreateRichMenuRequest{
        Size: api.RichMenuSize{
            Width:  2500,
            Height: height,
        },
        Selected:    false,
        Name:        name,
        ChatBarText: name,
        Areas:       areas,
    }

    // ... rest of function unchanged
```

**Step 3: Build and verify**

Run: `go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/cmd/richmenu.go
git commit -m "feat: add --size flag for compact rich menus"
```

---

## Task 5: Remove Unused DataBaseURL Constant

**Files:**
- Modify: `internal/api/client.go:13-16`

**Step 1: Remove unused constant**

Remove the `DataBaseURL` constant since it's not used:

```go
// Replace lines 13-16 with just:
const BaseURL = "https://api.line.me"
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/api/client.go
git commit -m "chore: remove unused DataBaseURL constant"
```

---

## Task 6: Remove Unused Mutex in Auth Server

**Files:**
- Modify: `internal/auth/server.go:25-31`

**Step 1: Remove unused mutex field**

Remove the `mu sync.Mutex` field from SetupServer struct:

```go
// Replace lines 25-31 with:
type SetupServer struct {
    result    chan SetupResult
    shutdown  chan struct{}
    csrfToken string
    store     secrets.Store
}
```

**Step 2: Remove sync import if unused**

Check if `sync` package is still needed. If not, remove from imports.

**Step 3: Build and verify**

Run: `go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/auth/server.go
git commit -m "chore: remove unused mutex from SetupServer"
```

---

## Task 7: Add --alt-text Flag for Flex Messages

**Files:**
- Modify: `internal/cmd/message.go`

**Step 1: Add alt-text flag to push command**

Update `newMessagePushCmd` to include `--alt-text` flag:

```go
// In newMessagePushCmd, add variable and flag:
var altText string

// Add flag after other flags:
cmd.Flags().StringVar(&altText, "alt-text", "Flex message", "Alt text for flex messages (shown in notifications)")
```

**Step 2: Update pushFlexMessage call**

Pass altText to the function:

```go
// Update the call in RunE:
return pushFlexMessage(cmd, userID, flexJSON, altText)
```

**Step 3: Update pushFlexMessage function signature**

```go
func pushFlexMessage(cmd *cobra.Command, userID, flexJSON, altText string) error {
    // ...
    if err := client.PushFlexMessage(cmd.Context(), userID, altText, json.RawMessage(flexJSON)); err != nil {
```

**Step 4: Do the same for broadcast command**

Add `--alt-text` flag to `newMessageBroadcastCmd` and update `broadcastFlexMessage` similarly.

**Step 5: Build and verify**

Run: `go build ./...`
Expected: No errors

**Step 6: Commit**

```bash
git add internal/cmd/message.go
git commit -m "feat: add --alt-text flag for flex messages"
```

---

## Task 8: Add Basic Unit Tests

**Files:**
- Create: `internal/api/client_test.go`
- Create: `internal/cmd/message_test.go`

**Step 1: Create API client test file**

```go
// internal/api/client_test.go
package api

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestClient_Get(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Authorization") != "Bearer test-token" {
            t.Errorf("expected Bearer test-token, got %s", r.Header.Get("Authorization"))
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    }))
    defer server.Close()

    client := NewClient("test-token")
    client.baseURL = server.URL

    data, err := client.Get(context.Background(), "/test")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    expected := `{"status":"ok"}`
    if string(data) != expected {
        t.Errorf("expected %s, got %s", expected, string(data))
    }
}

func TestClient_APIError(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte(`{"message":"bad request"}`))
    }))
    defer server.Close()

    client := NewClient("test-token")
    client.baseURL = server.URL

    _, err := client.Get(context.Background(), "/test")
    if err == nil {
        t.Fatal("expected error, got nil")
    }
}
```

**Step 2: Run tests**

Run: `go test ./internal/api/...`
Expected: PASS

**Step 3: Create message command test file**

```go
// internal/cmd/message_test.go
package cmd

import (
    "bytes"
    "testing"
)

func TestMessageCmd_RequiresSubcommand(t *testing.T) {
    cmd := newMessageCmd()

    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetErr(buf)
    cmd.SetArgs([]string{})

    // Running without subcommand should show help, not error
    err := cmd.Execute()
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}

func TestMessagePushCmd_RequiresTo(t *testing.T) {
    cmd := NewRootCmd()

    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetErr(buf)
    cmd.SetArgs([]string{"message", "push", "--text", "hello"})

    err := cmd.Execute()
    if err == nil {
        t.Error("expected error for missing --to flag")
    }
}
```

**Step 4: Run all tests**

Run: `go test ./...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/api/client_test.go internal/cmd/message_test.go
git commit -m "test: add basic unit tests for API client and commands"
```

---

## Summary

| Task | Description | Type |
|------|-------------|------|
| 1 | Fix division by zero | Important |
| 2 | Add token validation | Important |
| 3 | Add broadcast confirmation | Important |
| 4 | Add rich menu size option | Important |
| 5 | Remove unused DataBaseURL | Minor |
| 6 | Remove unused mutex | Minor |
| 7 | Add --alt-text flag | Minor |
| 8 | Add basic unit tests | Important |

**Total: 8 tasks**
