# Truth Social Go Library

A Go library for interacting with Truth Social's API, providing functionality to authenticate, lookup users, and fetch posts. This library uses the same authentication approach as Stanford's Truthbrush project.

## Features

- 🔐 **OAuth Authentication** - Secure authentication using Truth Social's OAuth flow
- 👤 **User Lookup** - Search and retrieve user account information
- 📝 **Post Fetching** - Pull posts from any public user account
- 🚀 **Easy to Use** - Simple API with comprehensive examples
- 🛡️ **Anti-Detection** - Uses CycleTLS for browser-like requests
- 📊 **Beautiful Tables** - Formatted output using tablewriter for CLI display

## Installation

```bash
go get github.com/nicolas-martin/truthsocial-go@v1.0.0
```

## Usage as a Library

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/nicolas-martin/truthsocial-go/client"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Create client
    tsClient, err := client.NewClient(ctx, "your_username", "your_password")
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer tsClient.Close()

    // Look up a user
    account, err := tsClient.Lookup(ctx, "realDonaldTrump")
    if err != nil {
        log.Fatalf("Failed to lookup user: %v", err)
    }
    
    fmt.Printf("User: @%s (%s)\n", account.Username, account.DisplayName)
    fmt.Printf("Followers: %d\n", account.FollowersCount)

    // Fetch posts
    posts, err := tsClient.PullStatuses(ctx, "truthsocial", true, 10)
    if err != nil {
        log.Fatalf("Failed to fetch posts: %v", err)
    }
    
    fmt.Printf("Found %d posts\n", len(posts))
}
```

## Quick Start

### Using the Command Line Interface

The included `main.go` provides a full-featured CLI for the library:

```bash
# Show help
go run main.go -help

# Run demo mode (requires authentication)
go run main.go -username your_username -password your_password

# Look up a specific user
go run main.go -username your_username -password your_password -lookup realDonaldTrump

# Fetch posts from a user
go run main.go -username your_username -password your_password -posts truthsocial -limit 20

# Get JSON output
go run main.go -username your_username -password your_password -lookup truthsocial -format json
```

## Command Line Options

### Authentication
- `-username string` - Truth Social username
- `-password string` - Truth Social password
- Environment variables: `TRUTHSOCIAL_USERNAME`, `TRUTHSOCIAL_PASSWORD`

### Operations
- `-lookup string` - Look up a specific user account
- `-posts string` - Fetch posts from a specific user
- `-limit int` - Number of posts to fetch (default: 10)
- `-exclude-replies` - Exclude replies from posts (default: true)

### Output
- `-format string` - Output format: 'pretty' or 'json' (default: pretty)
- `-help` - Show help message

## Examples

### Basic User Lookup

```bash
go run main.go -username myuser -password mypass -lookup realDonaldTrump
```

Output:
```
🔐 Authenticating with Truth Social...
✅ Authentication successful!
🔍 Looking up user: realDonaldTrump

👤 Account Information:
┌──────────────┬─────────────────────────┐
│    FIELD     │          VALUE          │
├──────────────┼─────────────────────────┤
│ Username     │ @realDonaldTrump        │
│ Display Name │ Donald J. Trump         │
│ Account ID   │ 107780257626128497      │
│ Followers    │ 6.8M                    │
│ Posts        │ 4.2K                    │
│ Verified     │ ✅ Yes                  │
└──────────────┴─────────────────────────┘
```

### Fetch Posts with JSON Output

```bash
go run main.go -username myuser -password mypass -posts truthsocial -limit 5 -format json
```
