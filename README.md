# 💬 LINE Official CLI — Your LINE bot, from the terminal.

Manage your LINE Official Account from the command line. Send messages, create rich menus, track analytics, manage audiences, and configure LIFF apps — all without leaving your terminal.

## Features

- **Audiences** - create and manage audience groups for targeted messaging
- **Authentication** - secure keychain storage, multi-account support
- **Bot Management** - get bot info, user profiles, follower lists
- **Chat Features** - loading animations, mark messages as read
- **Content** - download images, videos, and audio from messages
- **Coupons** - create, list, and manage promotional coupons
- **Groups & Rooms** - manage group chats and multi-person rooms
- **Insights** - view follower stats, message delivery, demographics
- **LIFF Apps** - create and manage LINE Front-end Framework apps
- **Memberships** - manage subscription plans and members (Japan)
- **Messaging** - push, broadcast, multicast, reply, narrowcast
- **Modules** - LINE Official Account Manager integration
- **PNP Messages** - send notifications by phone number (no LINE ID needed)
- **Rich Menus** - create, upload images, set defaults, bulk operations
- **Shop** - send mission stickers as rewards
- **Tokens** - issue, verify, and revoke channel access tokens
- **Webhooks** - configure endpoints, test connectivity, local dev server

## Installation

### Homebrew

```bash
brew install salmonumbrella/tap/line-official-cli
```

### From Source

```bash
go install github.com/salmonumbrella/line-official-cli/cmd/line@latest
```

## Quick Start

### 1. Get Your Channel Access Token

1. Go to [LINE Developers Console](https://developers.line.biz/console/)
2. Select your provider and Messaging API channel
3. Under "Messaging API", issue a long-lived channel access token

### 2. Authenticate

Choose one of two methods:

**Browser (recommended):**
```bash
line auth login
```

**Terminal:**
```bash
line auth login --token YOUR_TOKEN --name my-account
```

### 3. Test Your Setup

```bash
line bot info
```

## Configuration

### Account Selection

Specify the account using either a flag or environment variable:

```bash
# Via flag
line message push --account my-account --to USER_ID --text "Hello!"

# Via environment
export LINE_ACCOUNT=my-account
line message push --to USER_ID --text "Hello!"
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `LINE_ACCOUNT` | Default account name to use |
| `LINE_OUTPUT` | Output format: `text` (default) or `json` |

## Security

### Credential Storage

Credentials are stored securely in your system's keychain:

- **macOS**: Keychain Access
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **Windows**: Credential Manager
- **Fallback**: Encrypted file at `~/.line-cli/credentials`

## Commands

### Authentication

```bash
line auth login                        # Interactive login (opens browser)
line auth login --token TOKEN --name N # Login with token directly
line auth logout --name my-account     # Remove stored credentials
line auth status                       # Show current account
line auth list                         # List configured accounts
```

### Bot Management

```bash
line bot info                          # Bot display name, user ID, settings
line bot profile --user USER_ID        # Get user profile
line bot followers                     # List follower IDs (first 100)
line bot followers --all               # Fetch all followers (paginated)
line bot link-token --user USER_ID     # Generate account linking token
```

### Chat Features

```bash
# Show loading animation while processing
line chat loading --user USER_ID                 # Default 5 seconds
line chat loading --user USER_ID --seconds 10   # Custom duration (1-60)

# Mark messages as read
line chat mark-read --user USER_ID               # By user ID
line chat mark-read --token CHAT_TOKEN           # By webhook chat token
```

### Messaging

```bash
# Push to a single user
line message push --to USER_ID --text "Hello!"
line message push --to USER_ID --flex '{"type":"bubble",...}'
line message push --to USER_ID --image https://example.com/image.jpg
line message push --to USER_ID --sticker-package 446 --sticker-id 1988

# Broadcast to all followers (requires confirmation)
line message broadcast --text "Announcement!" --yes

# Multicast to multiple users (max 500)
line message multicast --to U123,U456,U789 --text "Hello group!"

# Reply to webhook event
line message reply --token REPLY_TOKEN --text "Thanks!"

# Targeted messaging
line message narrowcast --text "Special offer!" --audience 12345678
line message narrowcast-status --request-id REQUEST_ID

# Quota and stats
line message quota
line message delivery-stats --type broadcast --date 20251230
line message validate --type push --messages '[{"type":"text","text":"Hello"}]'
```

### Rich Menus

```bash
# List and manage
line richmenu list
line richmenu get --id richmenu-xxx
line richmenu delete --id richmenu-xxx

# Create with actions
line richmenu create --name "Main Menu" --size full \
  --actions '[{"type":"message","label":"Help","text":"help"}]'

# Upload image (2500x1686 for full, 2500x843 for compact)
line richmenu upload-image --id richmenu-xxx --image menu.png
line richmenu download-image --id richmenu-xxx

# Set default for all users
line richmenu set-default --id richmenu-xxx
line richmenu cancel-default

# Link to specific users
line richmenu link --user USER_ID --id richmenu-xxx
line richmenu unlink --user USER_ID

# Bulk operations
line richmenu bulk link --menu richmenu-xxx --users users.txt
line richmenu bulk unlink --users users.txt

# Aliases for human-readable references
line richmenu alias create --alias main-menu --id richmenu-xxx
line richmenu alias list
line richmenu alias get --alias main-menu
line richmenu alias update --alias main-menu --id richmenu-yyy
line richmenu alias delete --alias main-menu

# Batch operations (atomic)
line richmenu batch --operations ops.json
line richmenu batch status --request REQUEST_ID
line richmenu batch validate --operations ops.json

# Validation
line richmenu validate --file menu.json
```

### Audiences

```bash
# List and manage
line audience list
line audience get --id 12345678
line audience delete --id 12345678

# Create from user IDs
line audience create --name "VIP Users" --users U123,U456,U789
line audience create --name "Campaign Target" --file users.txt

# Add users to existing audience
line audience add-users --id 12345678 --users U123,U456

# Create from message interactions
line audience create-click --name "Clicked Link" --request REQUEST_ID
line audience create-impression --name "Saw Message" --request REQUEST_ID

# Update description
line audience update-description --id 12345678 --description "Updated name"

# Shared audiences
line audience shared list
line audience shared get --id 12345678
```

### Insights & Analytics

```bash
# Follower stats
line insight followers
line insight followers --date 20251230

# Message delivery stats
line insight messages
line insight messages --date 20251230

# Demographics (requires 20+ friends)
line insight demographics

# Message event stats
line insight events --request-id REQUEST_ID

# Stats per aggregation unit
line insight unit-stats --unit campaign-2024 --from 20251224 --to 20251231
```

### LIFF Apps

```bash
# List apps
line liff list

# Create (compact, tall, or full)
line liff create --type full --url https://example.com/liff --description "My App"

# Update
line liff update --id LIFF_ID --type tall --url https://example.com/new-url

# Delete
line liff delete --id LIFF_ID --yes
```

### Memberships (Japan)

```bash
# List membership plans
line membership plans

# Check user's subscription status
line membership status --user USER_ID

# List all membership subscribers
line membership users
line membership users --all              # Fetch all (paginated)
```

### Modules

```bash
# List bots with attached modules
line module bots

# Detach module from a LINE Official Account
line module detach --bot-id BOT_USER_ID --yes

# Chat control for module channels
line module acquire --chat USER_ID       # Take control from Primary Channel
line module acquire --chat USER_ID --no-expiry
line module release --chat USER_ID       # Return control to Primary Channel

# Exchange authorization code for module token
line module token --code AUTH_CODE --redirect-uri URI \
  --client-id ID --client-secret SECRET
```

### PNP Messages (Phone Number Push)

Send LINE notification messages to users by phone number instead of LINE user ID.
Requires a PNP-enabled channel.

```bash
# Send message by phone number (with country code)
line pnp push --to +819012345678 --text "Hello from LINE!"
```

### Webhooks

```bash
# Endpoint management
line webhook get                                 # Show current endpoint
line webhook set --url https://example.com/hook # Set endpoint
line webhook test                                # Test current endpoint
line webhook test --url https://example.com/hook # Test specific URL

# Local development server
line webhook serve                               # Start on port 8080
line webhook serve --port 9000                   # Custom port
line webhook serve --secret CHANNEL_SECRET       # Validate signatures
line webhook serve --forward http://localhost:3000/webhook  # Forward to app
line webhook serve --quiet                       # Only show errors
```

### Groups & Rooms

```bash
# Group management
line group summary --id GROUP_ID
line group members --id GROUP_ID
line group members --id GROUP_ID --all
line group member-profile --id GROUP_ID --user USER_ID
line group leave --id GROUP_ID --yes

# Room management (multi-person chats)
line room members --id ROOM_ID
line room members --id ROOM_ID --all
line room profile --id ROOM_ID --user USER_ID
line room leave --id ROOM_ID --yes
```

### Content Download

```bash
line content download --message-id MESSAGE_ID
line content download --message-id MESSAGE_ID --output image.jpg
line content preview --message-id MESSAGE_ID
line content status --message-id MESSAGE_ID
```

### Coupons

```bash
# List and manage
line coupon list
line coupon list --status running       # Filter by status
line coupon get --id COUPON_ID

# Create a coupon
line coupon create --title "Summer Sale" \
  --start 1704067200000 --end 1735689600000 \
  --max-use 1 --visibility PUBLIC --acquisition normal \
  --discount 500

# Close (discontinue) a coupon
line coupon close --id COUPON_ID
```

### Channel Access Tokens

```bash
# v2 tokens
line token issue --client-id ID --client-secret SECRET
line token verify --token TOKEN
line token revoke --token TOKEN

# v2.1 JWT-based tokens
line token issue-jwt --jwt JWT_ASSERTION
line token verify-jwt --token TOKEN
line token revoke-jwt --token TOKEN --client-id ID --client-secret SECRET
line token list-keys --jwt JWT_ASSERTION

# v3 stateless tokens (15-minute expiry, cannot be revoked)
line token issue-stateless --client-id ID --client-secret SECRET
```

### Shop Integration

```bash
line shop mission --to USER_ID --product-id 12345 --product-type STICKER
line shop mission --to USER_ID --product-id 12345 --product-type STICKER --send-message
```

## Output Formats

### Text

Human-readable output with formatting:

```bash
$ line bot info
Display Name: My Bot
User ID:      U1234567890abcdef
Basic ID:     @mybot
Chat Mode:    bot

$ line richmenu list
Rich Menus:
* richmenu-abc123  Main Menu (default)
  richmenu-def456  Secondary Menu
```

### JSON

Machine-readable output for scripting:

```bash
$ line bot info --output json
{
  "displayName": "My Bot",
  "userId": "U1234567890abcdef",
  "basicId": "@mybot",
  "chatMode": "bot"
}
```

### Table

Tabular output for lists:

```bash
$ line audience list --output table
ID          DESCRIPTION    STATUS    USERS    CREATED
12345678    VIP Users      READY     150      2025-01-15
12345679    Campaign       READY     500      2025-01-20
```

Data goes to stdout, errors and progress to stderr for clean piping.

## Examples

### Send a Flex Message

```bash
line message push --to USER_ID --flex '{
  "type": "bubble",
  "body": {
    "type": "box",
    "layout": "vertical",
    "contents": [
      {"type": "text", "text": "Hello from CLI!"}
    ]
  }
}'
```

### Create a Rich Menu with Image

```bash
# Create the menu
MENU_ID=$(line richmenu create --name "Main" --size full \
  --actions '[{"type":"message","label":"Help","text":"help"}]' \
  --output json | jq -r '.richMenuId')

# Upload the image
line richmenu upload-image --id "$MENU_ID" --image menu.png

# Set as default
line richmenu set-default --id "$MENU_ID"
```

### Multi-Account Usage

```bash
# Configure multiple accounts
line auth login --token TOKEN1 --name production
line auth login --token TOKEN2 --name staging

# Use specific account
line --account production message broadcast --text "Live!" --yes
line --account staging bot info

# Or set default via environment
export LINE_ACCOUNT=production
line message quota
```

### Automation with JSON Output

```bash
# Get follower count as JSON
line insight followers --output json | jq '.followers'

# List rich menus and extract default
line richmenu list --output json | jq '.defaultRichMenu'

# Pipeline: delete all non-default rich menus
line richmenu list --output json | \
  jq -r '.richmenus[] | select(.richMenuId != .defaultRichMenu) | .richMenuId' | \
  xargs -I{} line richmenu delete --id {} --yes
```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
line --debug message push --to USER_ID --text "Test"
# Shows: -> POST /v2/bot/message/push
# Shows: <- 200 OK (123ms)
```

### Dry-Run Mode

Preview what would be sent without actually sending:

```bash
line --dry-run message broadcast --text "Hello everyone!"
# Output:
# [DRY-RUN] Would broadcast message
# Type: text
# Content: Hello everyone!
# No message was sent (dry-run mode)
```

## Global Flags

All commands support these flags:

| Flag | Description |
|------|-------------|
| `--account <name>` | Account to use (overrides LINE_ACCOUNT) |
| `--output <format>` | Output format: `text`, `json`, or `table` |
| `--debug` | Enable debug output (shows API requests/responses) |
| `--dry-run` | Preview without executing (for mutations) |
| `--yes`, `-y` | Skip confirmation prompts (useful for scripts) |
| `--help` | Show help for any command |

## Shell Completions

Generate shell completions for your preferred shell:

### Bash

```bash
# macOS (Homebrew):
line completion bash > $(brew --prefix)/etc/bash_completion.d/line

# Linux:
line completion bash > /etc/bash_completion.d/line

# Or source directly:
source <(line completion bash)
```

### Zsh

```zsh
# Save to fpath:
line completion zsh > "${fpath[1]}/_line"

# Or add to .zshrc:
echo 'source <(line completion zsh)' >> ~/.zshrc
```

### Fish

```fish
line completion fish > ~/.config/fish/completions/line.fish
```

### PowerShell

```powershell
# Load for current session:
line completion powershell | Out-String | Invoke-Expression

# Or add to profile:
line completion powershell >> $PROFILE
```

## License

MIT

## Links

- [LINE Messaging API Documentation](https://developers.line.biz/en/docs/messaging-api/)
- [LINE Developers Console](https://developers.line.biz/console/)
