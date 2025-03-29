// server.go
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "os"

    "golang.org/x/sys/windows"
    "unsafe"
)

// KeyEvent represents a minimal keystroke event.
type KeyEvent struct {
    Char string `json:"char"` // e.g., "A", "b", "#", etc.
}

func main() {
    // Pick a port or use an environment variable.
    port := "9090"
    addr := ":" + port

    ln, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatalf("Failed to listen on %s: %v", addr, err)
    }
    defer ln.Close()

    log.Printf("Server listening on %s\n", addr)

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Printf("Accept error: %v\n", err)
            continue
        }
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    log.Printf("New connection from %s\n", conn.RemoteAddr().String())

    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        line := scanner.Bytes()
        var evt KeyEvent
        if err := json.Unmarshal(line, &evt); err != nil {
            log.Printf("JSON unmarshal error: %v\n", err)
            continue
        }
        if evt.Char != "" {
            log.Printf("Injecting character: %s\n", evt.Char)
            simulateKeystroke(evt.Char)
        }
    }

    if err := scanner.Err(); err != nil {
        log.Printf("Scanner error: %v\n", err)
    }
    log.Printf("Connection closed: %s\n", conn.RemoteAddr().String())
}

// simulateKeystroke sends a single character to the active window using SendInput.
// This is a *very* naive approach—doesn't handle shift keys or special layouts properly.
func simulateKeystroke(char string) {
    // Convert the first rune of the string to a Windows virtual-key if possible.
    // We do a simplistic approach here for ASCII characters.
    if len(char) < 1 {
        return
    }
    r := rune(char[0]) // take the first byte

    // We can map an ASCII code to a scan code or virtual key. For demonstration,
    // let's assume everything is just a "key down / key up" of the same character.
    // In reality, you'd do a more robust map from a character to a keycode/scan code.
    vk := windows.VK(r)

    // Prepare two INPUT structures: KEYDOWN and KEYUP.
    inputs := make([]windows.INPUT, 2)

    // KEYDOWN
    inputs[0].Type = windows.INPUT_KEYBOARD
    inputs[0].Ki = windows.KEYBDINPUT{
        Vk:         vk,
        Scan:       0,
        Time:       0,
        Flags:      0, // KEYEVENTF_KEYUP = 2, but here 0 = key down
        DwExtraInfo: 0,
    }

    // KEYUP
    inputs[1].Type = windows.INPUT_KEYBOARD
    inputs[1].Ki = windows.KEYBDINPUT{
        Vk:         vk,
        Scan:       0,
        Time:       0,
        Flags:      windows.KEYEVENTF_KEYUP,
        DwExtraInfo: 0,
    }

    // Send the combined input array. 
    // https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-sendinput
    // windows.SendInput(numberOfEvents, *arrayPointer, sizeOfINPUT)
    _, _, err := windows.SendInput(uint32(len(inputs)), unsafe.Pointer(&inputs[0]), int32(unsafe.Sizeof(inputs[0])))
    if err != nil {
        log.Printf("SendInput error: %v\n", err)
    }
}
