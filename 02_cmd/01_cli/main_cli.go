// client.go
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "os"
    "strings"
)

// KeyEvent is the same minimal structure we used on the server.
type KeyEvent struct {
    Char string `json:"char"`
}

func main() {
    // Typically you'd get this from a flag or env variable, but let's hardcode for demo:
    serverAddr := "127.0.0.1:9090"

    conn, err := net.Dial("tcp", serverAddr)
    if err != nil {
        log.Fatalf("Failed to connect to server: %v\n", err)
    }
    defer conn.Close()

    log.Printf("Connected to server at %s\n", serverAddr)
    log.Printf("Type characters here and press Enter. They will be sent to the server.\n")

    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("> ")
        if !scanner.Scan() {
            // EOF or error
            break
        }
        text := scanner.Text()
        // For demonstration, let's just send each character individually,
        // so we see them appear as single keystrokes on the server side.
        for _, r := range text {
            evt := KeyEvent{Char: string(r)}
            data, _ := json.Marshal(evt)
            // We'll write a line-based JSON message. 
            conn.Write(append(data, '\n'))
        }
        // Optionally, send a newline keystroke too:
        if strings.TrimSpace(text) != "" {
            evt := KeyEvent{Char: "\r"} // or "\n"
            data, _ := json.Marshal(evt)
            conn.Write(append(data, '\n'))
        }
    }
    if err := scanner.Err(); err != nil {
        log.Printf("Input error: %v\n", err)
    }
    log.Println("Client shutting down.")
}
