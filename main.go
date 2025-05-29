package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"truthsocial-go/internal/client"

	"github.com/olekukonko/tablewriter"
)

func main() {
	// Command line flags
	var (
		username       = flag.String("username", "", "Truth Social username for authentication")
		password       = flag.String("password", "", "Truth Social password for authentication")
		lookupUser     = flag.String("lookup", "", "Username to lookup (e.g., 'realDonaldTrump')")
		pullPosts      = flag.String("posts", "", "Username to pull posts from")
		limit          = flag.Int("limit", 10, "Number of posts to fetch (default: 10)")
		excludeReplies = flag.Bool("exclude-replies", true, "Exclude replies from posts (default: true)")
		outputFormat   = flag.String("format", "pretty", "Output format: 'pretty' or 'json'")
		help           = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Check for required authentication credentials
	if *username == "" || *password == "" {
		// Try to get from environment variables
		if envUser := os.Getenv("TRUTHSOCIAL_USERNAME"); envUser != "" {
			*username = envUser
		}
		if envPass := os.Getenv("TRUTHSOCIAL_PASSWORD"); envPass != "" {
			*password = envPass
		}

		if *username == "" || *password == "" {
			fmt.Println("Error: Username and password are required")
			fmt.Println("Provide them via flags or set TRUTHSOCIAL_USERNAME and TRUTHSOCIAL_PASSWORD environment variables")
			showHelp()
			os.Exit(1)
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize Truth Social client
	fmt.Println("🔐 Authenticating with Truth Social...")
	tsClient, err := client.NewClient(ctx, *username, *password)
	if err != nil {
		log.Fatalf("Failed to create Truth Social client: %v", err)
	}
	defer tsClient.Close()

	fmt.Println("✅ Authentication successful!")

	// Handle different operations based on flags
	switch {
	case *lookupUser != "":
		handleUserLookup(ctx, tsClient, *lookupUser, *outputFormat)
	case *pullPosts != "":
		handlePullPosts(ctx, tsClient, *pullPosts, *limit, *excludeReplies, *outputFormat)
	default:
		// Default demo mode - show various features
		runDemo(ctx, tsClient, *outputFormat)
	}
}

func handleUserLookup(ctx context.Context, tsClient *client.Client, username string, format string) {
	fmt.Printf("🔍 Looking up user: %s\n", username)

	account, err := tsClient.Lookup(ctx, username)
	if err != nil {
		log.Fatalf("Failed to lookup user %s: %v", username, err)
	}

	if format == "json" {
		data, _ := json.MarshalIndent(account, "", "  ")
		fmt.Println(string(data))
	} else {
		printAccountInfo(account)
	}
}

func handlePullPosts(ctx context.Context, tsClient *client.Client, username string, limit int, excludeReplies bool, format string) {
	fmt.Printf("📝 Fetching %d posts from %s (exclude replies: %v)\n", limit, username, excludeReplies)

	statuses, err := tsClient.PullStatuses(ctx, username, excludeReplies, limit)
	if err != nil {
		log.Fatalf("Failed to pull posts from %s: %v", username, err)
	}

	if format == "json" {
		data, _ := json.MarshalIndent(statuses, "", "  ")
		fmt.Println(string(data))
	} else {
		printStatuses(statuses)
	}
}

func runDemo(ctx context.Context, tsClient *client.Client, format string) {
	fmt.Println("\n🚀 Running Truth Social Go Library Demo")
	fmt.Println("=====================================")

	// Demo 1: Look up a popular account
	demoUsername := "realDonaldTrump"
	fmt.Printf("\n1️⃣  Looking up user: %s\n", demoUsername)

	account, err := tsClient.Lookup(ctx, demoUsername)
	if err != nil {
		fmt.Printf("❌ Failed to lookup %s: %v\n", demoUsername, err)
		// Try alternative username
		demoUsername = "truthsocial"
		fmt.Printf("Trying alternative username: %s\n", demoUsername)
		account, err = tsClient.Lookup(ctx, demoUsername)
		if err != nil {
			fmt.Printf("❌ Failed to lookup %s: %v\n", demoUsername, err)
			return
		}
	}

	if format == "json" {
		data, _ := json.MarshalIndent(account, "", "  ")
		fmt.Println(string(data))
	} else {
		printAccountInfo(account)
	}

	// Demo 2: Fetch recent posts
	fmt.Printf("\n2️⃣  Fetching recent posts from %s\n", account.Username)

	statuses, err := tsClient.PullStatuses(ctx, account.Username, true, 5)
	if err != nil {
		fmt.Printf("❌ Failed to fetch posts: %v\n", err)
		return
	}

	if format == "json" {
		data, _ := json.MarshalIndent(statuses, "", "  ")
		fmt.Println(string(data))
	} else {
		printStatuses(statuses)
	}

	fmt.Println("\n✨ Demo completed successfully!")
}

func printAccountInfo(account *client.Account) {
	fmt.Println("\n👤 Account Information:")

	table := tablewriter.NewTable(os.Stdout)
	table.Header([]string{"Field", "Value"})

	// Add account data rows
	table.Append([]string{"Username", "@" + account.Username})
	table.Append([]string{"Display Name", account.DisplayName})
	table.Append([]string{"Account ID", account.ID})
	table.Append([]string{"Followers", formatNumber(account.FollowersCount)})
	table.Append([]string{"Posts", formatNumber(account.StatusesCount)})

	verifiedStatus := "❌ No"
	if account.Verified {
		verifiedStatus = "✅ Yes"
	}
	table.Append([]string{"Verified", verifiedStatus})

	table.Render()
}

func printStatuses(statuses []client.Status) {
	fmt.Printf("\n📋 Found %d posts:\n", len(statuses))

	table := tablewriter.NewTable(os.Stdout)
	table.Header([]string{"#", "Author", "Posted", "Content", "Engagement"})

	for i, status := range statuses {
		// Clean and truncate content
		content := cleanContent(status.Content)
		if len(content) > 100 {
			content = content[:100] + "..."
		}

		// Format engagement stats
		engagement := fmt.Sprintf("💬 %d | 🔄 %d | ❤️ %d",
			status.RepliesCount, status.ReblogsCount, status.FavouritesCount)

		table.Append([]string{
			fmt.Sprintf("%d", i+1),
			"@" + status.Account.Username,
			formatTime(status.CreatedAt),
			content,
			engagement,
		})
	}

	table.Render()
}

func cleanContent(content string) string {
	// Remove HTML tags (basic cleanup)
	content = strings.ReplaceAll(content, "<p>", "")
	content = strings.ReplaceAll(content, "</p>", "\n")
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")

	// Remove other common HTML tags
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")

	// Clean up whitespace
	content = strings.TrimSpace(content)
	lines := strings.Split(content, "\n")
	var cleanLines []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}

	return strings.Join(cleanLines, " ")
}

func formatTime(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}

	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}

func formatNumber(num int) string {
	if num < 1000 {
		return strconv.Itoa(num)
	} else if num < 1000000 {
		return fmt.Sprintf("%.1fK", float64(num)/1000)
	} else {
		return fmt.Sprintf("%.1fM", float64(num)/1000000)
	}
}

func showHelp() {
	fmt.Println("Truth Social Go Library - Command Line Interface")
	fmt.Println("===============================================")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run main.go [OPTIONS]")
	fmt.Println()
	fmt.Println("AUTHENTICATION:")
	fmt.Println("  -username string    Truth Social username")
	fmt.Println("  -password string    Truth Social password")
	fmt.Println("  (or set TRUTHSOCIAL_USERNAME and TRUTHSOCIAL_PASSWORD env vars)")
	fmt.Println()
	fmt.Println("OPERATIONS:")
	fmt.Println("  -lookup string      Look up a specific user account")
	fmt.Println("  -posts string       Fetch posts from a specific user")
	fmt.Println("  -limit int          Number of posts to fetch (default: 10)")
	fmt.Println("  -exclude-replies    Exclude replies from posts (default: true)")
	fmt.Println()
	fmt.Println("OUTPUT:")
	fmt.Println("  -format string      Output format: 'pretty' or 'json' (default: pretty)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Run demo mode")
	fmt.Println("  go run main.go -username myuser -password mypass")
	fmt.Println()
	fmt.Println("  # Look up a specific user")
	fmt.Println("  go run main.go -username myuser -password mypass -lookup realDonaldTrump")
	fmt.Println()
	fmt.Println("  # Fetch posts from a user")
	fmt.Println("  go run main.go -username myuser -password mypass -posts truthsocial -limit 20")
	fmt.Println()
	fmt.Println("  # Get JSON output")
	fmt.Println("  go run main.go -username myuser -password mypass -lookup truthsocial -format json")
	fmt.Println()
	fmt.Println("  # Using environment variables")
	fmt.Println("  export TRUTHSOCIAL_USERNAME=myuser")
	fmt.Println("  export TRUTHSOCIAL_PASSWORD=mypass")
	fmt.Println("  go run main.go -posts realDonaldTrump")
}
