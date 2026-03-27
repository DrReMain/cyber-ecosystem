package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	jwt2 "github.com/golang-jwt/jwt/v5"
)

const (
	Version = "v1.0.0"
)

var supportedMethods = []string{"HS256", "HS384", "HS512"}

func main() {
	// If no arguments or first arg is not a known command, run interactive mode
	if len(os.Args) < 2 {
		runInteractive()
		return
	}

	switch os.Args[1] {
	case "generate", "gen":
		generateToken()
	case "verify", "vrf":
		verifyToken()
	case "interactive", "i", "shell":
		runInteractive()
	case "help", "--help", "-h":
		printUsage()
	default:
		// Unknown command, show help
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`JWT Tool for Cyber Ecosystem

Usage:
  go-jwt [command] [flags]
  go-jwt interactive      Enter interactive mode

Commands:
  generate, gen        Generate a JWT token
  verify, vrf          Verify a JWT token
  interactive, i       Enter interactive mode (default)
  help                 Show this help message

Generate Token:
  go-jwt generate --secret <secret> --sub <subject> [--method <HS256|HS384|HS512>] [--exp <duration>] [--claims <json>]

  --secret string    Secret key for signing (required)
  --sub string       Subject claim - user identifier (required)
  --method string    Signing method: HS256, HS384, HS512 (default: HS256)
  --exp duration     Token expiration time (default: 24h)
                     Examples: 30m, 2h, 7d, 30d, 12w
  --claims string    Additional claims in JSON format (optional)
                     Example: '{"role":"admin","permissions":["read","write"]}'

Verify Token:
  go-jwt verify --secret <secret> --token <token>

  --secret string    Secret key for verification (required)
  --token string     JWT token to verify (required)

Examples:
  # Generate a token with default method (HS256) and expiration (24h)
  go-jwt generate --secret my-secret-key --sub user123

  # Generate a token with 30 days expiration
  go-jwt generate --secret my-secret-key --sub user123 --exp 30d

  # Generate a token with HS384 and 12 weeks expiration
  go-jwt generate --secret my-secret-key --sub user123 --method HS384 --exp 12w

  # Generate a token with additional claims
  go-jwt generate --secret my-secret-key --sub user123 --claims '{"role":"admin"}'

  # Verify a token
  go-jwt verify --secret my-secret-key --token eyJhbGciOiJIUzI1NiIs...

  # Enter interactive mode
  go-jwt interactive

Duration format:
  30m  = 30 minutes
  2h   = 2 hours
  7d   = 7 days
  30d  = 30 days
  12w  = 12 weeks (approx 3 months)

Note:
  The secret key must match the 'auth.api_key' in config.yaml:
  - Development: "some-secret-key-for-forntend"
  - Production:  Should use environment variable or secure vault
`)
}

// parseDuration supports d (days) and w (weeks) in addition to standard Go durations
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)

	// Try standard Go duration first
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Handle days and weeks
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		if days, err := strconv.ParseFloat(daysStr, 64); err == nil {
			return time.Duration(days * 24 * float64(time.Hour)), nil
		}
	}

	if strings.HasSuffix(s, "w") {
		weeksStr := strings.TrimSuffix(s, "w")
		if weeks, err := strconv.ParseFloat(weeksStr, 64); err == nil {
			return time.Duration(weeks * 7 * 24 * float64(time.Hour)), nil
		}
	}

	return time.ParseDuration(s)
}

func runInteractive() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║     JWT Tool for Cyber Ecosystem     ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	for {
		fmt.Println("Select operation:")
		fmt.Println("  [1] Generate JWT token")
		fmt.Println("  [2] Verify JWT token")
		fmt.Println("  [Q] Quit")
		fmt.Print("\nEnter choice: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToUpper(choice))

		switch choice {
		case "1", "G", "GENERATE":
			generateTokenInteractive(reader)
		case "2", "V", "VERIFY":
			verifyTokenInteractive(reader)
		case "Q", "QUIT", "EXIT":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("\nInvalid choice. Please try again.\n")
		}
		fmt.Println()
	}
}

func generateTokenInteractive(reader *bufio.Reader) {
	fmt.Println("\n━━━ Generate JWT Token ━━━")

	// Select signing method
	fmt.Println("\nSupported signing methods:")
	for i, m := range supportedMethods {
		fmt.Printf("  [%d] %s\n", i+1, m)
	}
	fmt.Print("Select method (default: 1): ")
	methodChoice, _ := reader.ReadString('\n')
	methodChoice = strings.TrimSpace(methodChoice)

	methodIndex := 0
	if methodChoice != "" {
		fmt.Sscanf(methodChoice, "%d", &methodIndex)
		methodIndex--
	}
	if methodIndex < 0 || methodIndex >= len(supportedMethods) {
		methodIndex = 0
	}
	signingMethod := supportedMethods[methodIndex]
	fmt.Printf("Using: %s\n", signingMethod)

	// Get secret
	fmt.Print("\nEnter secret key: ")
	secret, _ := reader.ReadString('\n')
	secret = strings.TrimSpace(secret)
	if secret == "" {
		fmt.Println("Error: secret key is required")
		return
	}

	// Get subject
	fmt.Print("Enter subject (user identifier): ")
	subject, _ := reader.ReadString('\n')
	subject = strings.TrimSpace(subject)
	if subject == "" {
		fmt.Println("Error: subject is required")
		return
	}

	// Get expiration
	fmt.Print("Enter expiration time (e.g. 24h, 7d, 30d, 12w) [default: 24h]: ")
	expStr, _ := reader.ReadString('\n')
	expStr = strings.TrimSpace(expStr)
	if expStr == "" {
		expStr = "24h"
	}
	expDuration, err := parseDuration(expStr)
	if err != nil {
		fmt.Printf("Error: Invalid expiration duration '%s': %v\n", expStr, err)
		return
	}

	// Get additional claims
	fmt.Print("Enter additional claims (JSON format, e.g. {\"role\":\"admin\"}) [optional, press Enter to skip]: ")
	claimsJSON, _ := reader.ReadString('\n')
	claimsJSON = strings.TrimSpace(claimsJSON)

	var additionalClaims map[string]interface{}
	if claimsJSON != "" {
		if err := json.Unmarshal([]byte(claimsJSON), &additionalClaims); err != nil {
			fmt.Printf("Error: Invalid claims JSON: %v\n", err)
			return
		}
	}

	// Build claims
	now := time.Now()
	claims := jwt2.MapClaims{
		"sub": subject,
		"iat": now.Unix(),
		"exp": now.Add(expDuration).Unix(),
		"nbf": now.Unix(),
	}

	for k, v := range additionalClaims {
		claims[k] = v
	}

	// Create token
	var token *jwt2.Token
	switch signingMethod {
	case "HS256":
		token = jwt2.NewWithClaims(jwt2.SigningMethodHS256, claims)
	case "HS384":
		token = jwt2.NewWithClaims(jwt2.SigningMethodHS384, claims)
	case "HS512":
		token = jwt2.NewWithClaims(jwt2.SigningMethodHS512, claims)
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Printf("Error: Failed to sign token: %v\n", err)
		return
	}

	fmt.Println("\n✅ Generated JWT Token:")
	fmt.Println("─" + strings.Repeat("─", 40))
	fmt.Println(tokenString)
	fmt.Println("─" + strings.Repeat("─", 40))

	expTime := now.Add(expDuration)
	fmt.Printf("Expires: %s (%s from now)\n", expTime.Format(time.RFC3339), expDuration)
}

func verifyTokenInteractive(reader *bufio.Reader) {
	fmt.Println("\n━━━ Verify JWT Token ━━━")

	// Get secret
	fmt.Print("\nEnter secret key: ")
	secret, _ := reader.ReadString('\n')
	secret = strings.TrimSpace(secret)
	if secret == "" {
		fmt.Println("Error: secret key is required")
		return
	}

	// Get token
	fmt.Print("Enter JWT token: ")
	tokenString, _ := reader.ReadString('\n')
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		fmt.Println("Error: token is required")
		return
	}

	// Parse and validate token
	token, err := jwt2.Parse(tokenString, func(token *jwt2.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt2.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		fmt.Printf("\n❌ Token verification failed: %v\n", err)
		return
	}

	if !token.Valid {
		fmt.Println("\n❌ Token is invalid")
		return
	}

	// Extract claims
	claims, ok := token.Claims.(jwt2.MapClaims)
	if !ok {
		fmt.Println("\n❌ Failed to extract claims")
		return
	}

	fmt.Println("\n✅ Token Verified Successfully!")
	fmt.Println("\nClaims:")
	claimsJSON, err := json.MarshalIndent(claims, "", "  ")
	if err != nil {
		fmt.Printf("  %v\n", claims)
	} else {
		fmt.Printf("  %s\n", claimsJSON)
	}

	// Print human-readable expiration
	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		remaining := time.Until(expTime)
		if remaining > 0 {
			fmt.Printf("Expires: %s (%s remaining)\n", expTime.Format(time.RFC3339), remaining)
		} else {
			fmt.Printf("Expired: %s (%s ago)\n", expTime.Format(time.RFC3339), -remaining)
		}
	}

	// Print signing method used
	if token.Method != nil {
		fmt.Printf("Signing Method: %s\n", token.Method.Alg())
	}
}

func generateToken() {
	genCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	secret := genCmd.String("secret", "", "Secret key for signing")
	subject := genCmd.String("sub", "", "Subject claim - user identifier")
	signingMethod := genCmd.String("method", "HS256", "Signing method: HS256, HS384, HS512")
	expDuration := genCmd.String("exp", "24h", "Token expiration time")
	claimsJSON := genCmd.String("claims", "{}", "Additional claims in JSON format")

	genCmd.Parse(os.Args[2:])

	if *secret == "" {
		fmt.Println("Error: --secret is required")
		os.Exit(1)
	}
	if *subject == "" {
		fmt.Println("Error: --sub is required")
		os.Exit(1)
	}

	// Validate signing method
	method := strings.ToUpper(*signingMethod)
	validMethod := false
	for _, m := range supportedMethods {
		if m == method {
			validMethod = true
			break
		}
	}
	if !validMethod {
		fmt.Printf("Error: Invalid signing method '%s'. Supported: %v\n", *signingMethod, supportedMethods)
		os.Exit(1)
	}

	// Parse expiration duration (with support for d and w)
	expDurationParsed, err := parseDuration(*expDuration)
	if err != nil {
		fmt.Printf("Error: Invalid expiration duration '%s': %v\n", *expDuration, err)
		os.Exit(1)
	}

	// Parse additional claims
	var additionalClaims map[string]interface{}
	if err := json.Unmarshal([]byte(*claimsJSON), &additionalClaims); err != nil {
		fmt.Printf("Error: Invalid claims JSON: %v\n", err)
		os.Exit(1)
	}

	// Build claims
	now := time.Now()
	claims := jwt2.MapClaims{
		"sub": *subject,
		"iat": now.Unix(),
		"exp": now.Add(expDurationParsed).Unix(),
		"nbf": now.Unix(),
	}

	for k, v := range additionalClaims {
		claims[k] = v
	}

	// Create token based on method
	var token *jwt2.Token
	switch method {
	case "HS256":
		token = jwt2.NewWithClaims(jwt2.SigningMethodHS256, claims)
	case "HS384":
		token = jwt2.NewWithClaims(jwt2.SigningMethodHS384, claims)
	case "HS512":
		token = jwt2.NewWithClaims(jwt2.SigningMethodHS512, claims)
	}

	tokenString, err := token.SignedString([]byte(*secret))
	if err != nil {
		fmt.Printf("Error: Failed to sign token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Generated JWT Token ===")
	fmt.Println(tokenString)
	fmt.Println("===========================")
	fmt.Printf("Signing Method: %s\n", method)
	fmt.Printf("Expires in: %s\n", expDurationParsed)
}

func verifyToken() {
	vrfCmd := flag.NewFlagSet("verify", flag.ExitOnError)
	secret := vrfCmd.String("secret", "", "Secret key for verification")
	tokenString := vrfCmd.String("token", "", "JWT token to verify")

	vrfCmd.Parse(os.Args[2:])

	if *secret == "" {
		fmt.Println("Error: --secret is required")
		os.Exit(1)
	}
	if *tokenString == "" {
		fmt.Println("Error: --token is required")
		os.Exit(1)
	}

	// Parse and validate token (accept any HMAC method)
	token, err := jwt2.Parse(*tokenString, func(token *jwt2.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt2.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(*secret), nil
	})

	if err != nil {
		fmt.Printf("Error: Token verification failed: %v\n", err)
		os.Exit(1)
	}

	if !token.Valid {
		fmt.Println("Error: Token is invalid")
		os.Exit(1)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt2.MapClaims)
	if !ok {
		fmt.Println("Error: Failed to extract claims")
		os.Exit(1)
	}

	fmt.Println("=== Token Verified Successfully ===")

	// Print signing method
	if token.Method != nil {
		fmt.Printf("Signing Method: %s\n", token.Method.Alg())
	}

	// Pretty print claims
	claimsJSON, err := json.MarshalIndent(claims, "", "  ")
	if err != nil {
		fmt.Printf("Claims: %v\n", claims)
	} else {
		fmt.Printf("Claims:\n%s\n", claimsJSON)
	}

	// Print human-readable expiration
	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		remaining := time.Until(expTime)
		if remaining > 0 {
			fmt.Printf("Expires: %s (%s remaining)\n", expTime.Format(time.RFC3339), remaining)
		} else {
			fmt.Printf("Expired: %s (%s ago)\n", expTime.Format(time.RFC3339), -remaining)
		}
	}

	fmt.Println("=================================")
}
