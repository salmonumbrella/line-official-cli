# LINE Official CLI

A command-line interface for managing LINE Official Accounts. Built for humans and AI agents with 100% LINE Messaging API coverage.

## Features

- **Complete API Coverage** - 120+ endpoints covering messaging, rich menus, audiences, insights, LIFF, and more
- **Multi-Account Support** - Store and switch between multiple LINE Official Accounts
- **Secure Credentials** - Uses system keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- **Agent-Friendly** - JSON output mode and `--yes` flag for automation
- **Cross-Platform** - Works on macOS, Linux, and Windows

## Installation

### From Source

```bash
go install github.com/salmonumbrella/line-official-cli/cmd/line@latest
```

### Build Locally

```bash
git clone https://github.com/salmonumbrella/line-official-cli.git
cd line-official-cli
make build    # Creates ./line binary
make install  # Installs to $GOPATH/bin
```

## Quick Start

### 1. Get Your Channel Access Token

1. Go to [LINE Developers Console](https://developers.line.biz/console/)
2. Select your provider and Messaging API channel
3. Under "Messaging API", issue a long-lived channel access token

### 2. Authenticate

```bash
# Interactive login (opens browser)
line auth login

# Or provide token directly
line auth login --token YOUR_TOKEN --name my-account
```

### 3. Send Your First Message

```bash
# Push a message to a user
line message push --to USER_ID --text "Hello from LINE CLI!"

# Broadcast to all followers (requires confirmation)
line message broadcast --text "Hello everyone!"
```

## Command Reference

### Authentication

```bash
line auth login                    # Interactive login
line auth login --token TOKEN      # Login with token
line auth logout --name my-account # Remove stored credentials
line auth status                   # Show current account
line auth list                     # List configured accounts
```

### Messaging

```bash
# Push to a single user
line message push --to USER_ID --text "Hello!"
line message push --to USER_ID --flex '{"type":"bubble",...}'
line message push --to USER_ID --image https://example.com/image.jpg
line message push --to USER_ID --sticker-package 446 --sticker-id 1988

# Broadcast to all followers
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
line richmenu bulk link --menu richmenu-xxx --users users.txt

# Aliases for human-readable references
line richmenu alias create --alias main-menu --id richmenu-xxx
line richmenu alias list
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

### Webhooks

```bash
line webhook get                              # Show current endpoint
line webhook set --url https://example.com/webhook
line webhook test                             # Test current endpoint
line webhook test --url https://example.com/webhook  # Test specific URL
```

### Bot Management

```bash
line bot info                        # Bot display name, user ID, settings
line bot profile --user USER_ID      # Get user profile
line bot followers                   # List follower IDs
line bot followers --all             # Fetch all (paginated)
line bot link-token --user USER_ID   # Generate account linking token
```

### Groups & Rooms

```bash
# Group management
line group summary --id GROUP_ID
line group members --id GROUP_ID
line group leave --id GROUP_ID

# Room management
line room members --id ROOM_ID
line room leave --id ROOM_ID
```

### Additional Commands

```bash
# Content download
line content download --message-id MESSAGE_ID --output image.jpg

# Channel access tokens
line token issue --client-id ID --client-secret SECRET
line token verify --token TOKEN
line token revoke --token TOKEN

# Shop integration
line shop info
```

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `LINE_ACCOUNT` | Default account name to use |
| `LINE_OUTPUT` | Default output format (`text` or `json`) |

### Global Flags

| Flag | Description |
|------|-------------|
| `--account NAME` | Select which stored account to use |
| `--output FORMAT` | Output format: `text` (default) or `json` |
| `--yes`, `-y` | Skip confirmation prompts |
| `--debug` | Enable debug output |

### Credential Storage

Credentials are stored securely in your system's keychain:

- **macOS**: Keychain Access
- **Windows**: Credential Manager
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **Fallback**: Encrypted file at `~/.line-cli/credentials`

## API Coverage

This CLI provides 100% coverage of the LINE Messaging API, including:

- Messaging (push, broadcast, multicast, reply, narrowcast)
- Rich menus (CRUD, aliases, bulk operations, batch operations)
- Audience management (upload, click/impression-based, shared audiences)
- Insights (followers, messages, demographics, events)
- LIFF app management
- Webhook configuration
- Bot info and user profiles
- Group and room management
- Content download
- Channel access token management
- PNP (Push Notification Push) messaging
- Module channel operations

## Examples

### Automation with JSON Output

```bash
# Get follower count as JSON
line insight followers --output json | jq '.followers'

# List rich menus and get default
line richmenu list --output json | jq '.defaultRichMenu'

# Send message without confirmation
line message broadcast --text "Daily update" --yes
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

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests (`make test`)
4. Run linter (`make lint`)
5. Commit your changes
6. Push to the branch
7. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.
