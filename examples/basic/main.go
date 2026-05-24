package main

import (
	"context"
	"log"

	"github.com/example/go-diamond/pkg/diamond"
)

func main() {
	client := diamond.NewClient("http://127.0.0.1:8080", "default")

	client.AddListener("DEFAULT_GROUP", "app.json", func(content string) {
		log.Printf("config changed: %s", content)
	})

	ctx := context.Background()
	client.Start(ctx)

	select {}
}