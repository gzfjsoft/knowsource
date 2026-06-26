package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	AccountFile = ".knowsource-account.json"
	TokenFile   = ".knowsource-session_token"
	DefaultHost = "http://localhost:8070"
)

type AccountConfig struct {
	ClientID string `json:"clientId"`
	EmpCode  string `json:"empCode"`
	Password string `json:"password"`
	IsDebug  int64  `json:"isDebug"`
}

type AccountList struct {
	Accounts []AccountConfig `json:"accounts"`
	Default  string          `json:"default"` // "clientId:empCode"
}

func loadAccountList() AccountList {
	var list AccountList
	data, err := os.ReadFile(AccountFile)
	if err == nil {
		_ = json.Unmarshal(data, &list)
	}
	if list.Accounts == nil {
		list.Accounts = []AccountConfig{}
	}
	return list
}

func saveAccountList(list AccountList) {
	data, err := json.MarshalIndent(list, "", "  ")
	if err == nil {
		_ = os.WriteFile(AccountFile, data, 0600)
	}
}

func parseArgs(args []string) (string, map[string]string) {
	if len(args) == 0 {
		return "", nil
	}
	cmd := strings.ToLower(args[0])
	opts := make(map[string]string)
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			key := strings.TrimPrefix(arg, "--")
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				opts[key] = args[i+1]
				i++
			} else {
				opts[key] = "true"
			}
		} else if strings.HasPrefix(arg, "-") {
			key := strings.TrimPrefix(arg, "-")
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				opts[key] = args[i+1]
				i++
			} else {
				opts[key] = "true"
			}
		}
	}
	return cmd, opts
}

func getOption(opts map[string]string, keys ...string) string {
	for _, k := range keys {
		if v, ok := opts[k]; ok {
			return v
		}
	}
	return ""
}

func hasOption(opts map[string]string, keys ...string) bool {
	for _, k := range keys {
		if _, ok := opts[k]; ok {
			return true
		}
	}
	return false
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		return
	}

	cmd, opts := parseArgs(args)

	host := getOption(opts, "host")
	if host == "" {
		host = DefaultHost
	}

	client := NewClient(host)

	tokenBytes, err := os.ReadFile(TokenFile)
	if err == nil {
		token := strings.TrimSpace(string(tokenBytes))
		client.SetToken(token)
	}

	switch cmd {
	case "login":
		handleLogin(client, opts)
	case "documents-type-list", "doc-type-list":
		handleListDocumentsType(client, opts)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Knowsource CLI")
	fmt.Println("Usage: knowsource-cli <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  login                 Login to Knowsource")
	fmt.Println("  doc-type-list         List knowledge base document types")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --host <url>          Server base URL (default: http://localhost:8071)")
	fmt.Println("  --json                Output raw JSON response")
	fmt.Println()
	fmt.Println("Login Options:")
	fmt.Println("  -c, --client <id>     Client (tenant) ID")
	fmt.Println("  -u, --user <code>     Employee code")
	fmt.Println("  -p, --passwd <pwd>    Password")
	fmt.Println("  -d, --debug           Enable debug login mode")
}

func handleLogin(client *Client, opts map[string]string) {
	var clientID, empCode, password string
	var isDebug int64

	accountList := loadAccountList()

	providedClient := getOption(opts, "client", "c")
	providedUser := getOption(opts, "user", "u")
	providedPasswd := getOption(opts, "passwd", "p")
	providedDebug := hasOption(opts, "debug", "d")

	if providedClient != "" {
		clientID = providedClient
		empCode = providedUser
		password = providedPasswd
		if providedDebug {
			isDebug = 1
		}

		if password == "" {
			for _, acc := range accountList.Accounts {
				if acc.ClientID == clientID && acc.EmpCode == empCode {
					password = acc.Password
					isDebug = acc.IsDebug
					break
				}
			}
		}

		if password == "" {
			fmt.Println("Error: Password not provided.")
			return
		}
	} else if accountList.Default != "" {
		parts := strings.SplitN(accountList.Default, ":", 2)
		if len(parts) == 2 {
			defClient := parts[0]
			defUser := parts[1]
			for _, acc := range accountList.Accounts {
				if acc.ClientID == defClient && acc.EmpCode == defUser {
					clientID = acc.ClientID
					empCode = acc.EmpCode
					password = acc.Password
					isDebug = acc.IsDebug
					break
				}
			}
		}
	}

	if clientID == "" || empCode == "" || password == "" {
		if len(accountList.Accounts) > 0 {
			first := accountList.Accounts[0]
			clientID = first.ClientID
			empCode = first.EmpCode
			password = first.Password
			isDebug = first.IsDebug
		} else {
			fmt.Println("Error: Missing credentials.")
			fmt.Println("Usage: knowsource-cli login --client <clientId> --user <empCode> --passwd <password>")
			return
		}
	}

	fmt.Printf("Logging in to %s as %s:%s...\n", client.BaseURL, clientID, empCode)

	req := LoginRequest{
		ClientID:  clientID,
		EmpCode:   empCode,
		Password:  password,
		IsDebug:   isDebug,
		Captcha:   "xxxxxxxxxxxxxxxxxxxxxxxxx",
		CaptchaId: "cli",
	}

	resp, err := client.Login(req)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}

	if resp.Code != 200 {
		fmt.Printf("Login failed (Code %d): %s\n", resp.Code, resp.Message)
		return
	}

	if resp.Data == nil || resp.Data.Token == "" {
		fmt.Println("Login failed: No token received in response.")
		return
	}

	err = os.WriteFile(TokenFile, []byte(resp.Data.Token), 0600)
	if err != nil {
		fmt.Printf("Warning: failed to save token file: %v\n", err)
	}

	newAccounts := []AccountConfig{}
	for _, acc := range accountList.Accounts {
		if acc.ClientID == clientID && acc.EmpCode == empCode {
			continue
		}
		newAccounts = append(newAccounts, acc)
	}
	newAccounts = append([]AccountConfig{{
		ClientID: clientID,
		EmpCode:  empCode,
		Password: password,
		IsDebug:  isDebug,
	}}, newAccounts...)

	accountList.Accounts = newAccounts
	accountList.Default = clientID + ":" + empCode
	saveAccountList(accountList)

	fmt.Printf("Successfully logged in!\n")
	fmt.Printf("User: %s (%s)\n", resp.Data.UserInfo.EmpName, resp.Data.UserInfo.Position)
	fmt.Printf("Dept: %s\n", resp.Data.UserInfo.DeptName)
	fmt.Printf("Client: %s (%s)\n", resp.Data.UserInfo.ClientID, resp.Data.UserInfo.ClientName)
	fmt.Printf("Account saved to %s\n", AccountFile)
}

func handleListDocumentsType(client *Client, opts map[string]string) {
	if client.Token == "" {
		fmt.Println("Error: Not logged in. Run 'knowsource-cli login' first.")
		return
	}

	resp, err := client.ListDocumentsType()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if resp.Code != 200 {
		fmt.Printf("Error (Code %d): %s\n", resp.Code, resp.Message)
		return
	}

	useJSON := hasOption(opts, "json")
	if useJSON {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(data))
		return
	}

	if resp.Data == nil || len(resp.Data.List) == 0 {
		fmt.Println("No document types found.")
		return
	}

	fmt.Printf("Total: %d\n\n", resp.Data.Total)
	fmt.Println("| Code | Name | Description | Tags | IsDisabled |")
	fmt.Println("|---|---|---|---|---|")
	for _, item := range resp.Data.List {
		tagsStr := strings.Join(item.Tags, ", ")
		fmt.Printf("| %s | %s | %s | %s | %d |\n",
			item.Code,
			item.Name,
			item.Description,
			tagsStr,
			item.IsDisabled,
		)
	}
}
