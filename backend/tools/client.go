package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		Username  string `json:"username"`
		CreatedAt string `json:"created_at"`
	} `json:"user"`
}

type ViewCountResponse struct {
	Count int `json:"count"`
}

var authToken string

func authenticate() error {
	fmt.Print("Username: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	username := strings.TrimSpace(scanner.Text())

	fmt.Print("Password: ")
	// Read password without echoing it to the terminal
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("error reading password: %v", err)
	}
	fmt.Println() // Add a newline after password input

	password := string(passwordBytes)

	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("error preparing login request: %v", err)
	}

	resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: %s", body)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("error parsing login response: %v", err)
	}

	authToken = loginResp.Token
	fmt.Printf("Logged in as %s\n", loginResp.User.Username)
	return nil
}

func main() {
	fmt.Println("Welcome to the checklist client!")
	fmt.Println("Please log in to continue.")

	if err := authenticate(); err != nil {
		fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("$ ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		command, args := parts[0], parts[1:]

		switch command {
		case "exit":
			code := 0
			if len(args) > 0 {
				if n, err := strconv.Atoi(args[0]); err == nil {
					code = n
				}
			}
			os.Exit(code)

		case "views":
			getViewCount()

		default:
			fmt.Printf("%s: command not found\n", command)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}

func getViewCount() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8080/", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Request failed with status %d: %s\n", resp.StatusCode, body)
		return
	}

	var viewResp ViewCountResponse
	if err := json.NewDecoder(resp.Body).Decode(&viewResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		return
	}

	fmt.Printf("Current view count: %d\n", viewResp.Count)
}
