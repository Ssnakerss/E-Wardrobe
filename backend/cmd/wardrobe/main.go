package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "invite":
		handleInvite(os.Args[2:])
	case "recovery":
		handleRecovery(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: wardrobe <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  invite create    Create a new invitation token")
	fmt.Println("  invite revoke    Revoke an existing invitation")
	fmt.Println("  recovery create  Create a recovery link for a user")
	fmt.Println("  help             Show this help")
}

func handleInvite(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "invite requires a subcommand: create, revoke\n")
		os.Exit(1)
	}

	subcmd := args[0]

	switch subcmd {
	case "create":
		handleInviteCreate(args[1:])
	case "revoke":
		handleInviteRevoke(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown invite subcommand: %s\n", subcmd)
		os.Exit(1)
	}
}

func handleInviteCreate(args []string) {
	fs := flag.NewFlagSet("invite create", flag.ExitOnError)
	email := fs.String("email", "", "Email to bind the invitation to")
	expires := fs.Duration("expires", 7*24*time.Hour, "Expiration duration")

	fs.Parse(args)

	fmt.Printf("Invitation create command placeholder\n")
	fmt.Printf("  email: %s\n", *email)
	fmt.Printf("  expires: %s\n", *expires)
	// TODO: implement actual invitation creation
}

func handleInviteRevoke(args []string) {
	// TODO: implement invitation revocation
	fmt.Println("Invitation revoke command placeholder")
}

func handleRecovery(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "recovery requires a subcommand: create\n")
		os.Exit(1)
	}

	subcmd := args[0]

	switch subcmd {
	case "create":
		// TODO: implement recovery token creation
		fmt.Println("Recovery create command placeholder")
	default:
		fmt.Fprintf(os.Stderr, "unknown recovery subcommand: %s\n", subcmd)
		os.Exit(1)
	}
}
