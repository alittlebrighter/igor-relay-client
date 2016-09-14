package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now().Unix()
	fmt.Printf("Now: %v\n", now)
	fmt.Printf("Later: %v\n", now + (10 * time.Second))
}
