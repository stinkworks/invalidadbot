package main

import (
    // "github.com/go-telegram/bot"
    "os"
    "fmt"
    // "log"
)

func main() {
    path, exists := os.LookupEnv("PATH")

    if exists {
        fmt.Print(path)
    }
}
