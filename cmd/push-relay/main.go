package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aether/code_aether/pkg/v07/push"
)

var (
	pushProfile = flag.String("profile", "default", "push relay profile name")
	relayPort   = flag.String("port", "5005", "push relay listen port")
)

func main() {
	flag.Parse()
	if err := runPushRelay(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "push relay failed: %v\n", err)
		os.Exit(1)
	}
}

func runPushRelay(ctx context.Context) error {
	adapter := push.NewAdapter("mock", &push.MockAdapter{})
	service := push.NewService(adapter)
	authToken := fmt.Sprintf("relay-auth-%s", *pushProfile)
	reg := push.Registration{
		DeviceID:  fmt.Sprintf("device-%s", *pushProfile),
		ClientID:  fmt.Sprintf("client-%s", *pushProfile),
		Token:     authToken,
		CreatedAt: time.Now().UTC(),
		Metadata: map[string]string{
			"profile":            *pushProfile,
			push.AuthMetadataKey: authToken,
		},
	}
	if err := service.Register(ctx, reg); err != nil {
		return err
	}
	payload := push.BuildPayload(fmt.Sprintf("payload-%s", *pushProfile), reg.DeviceID, []byte("hello"), map[string]string{
		"source":             "cli",
		push.AuthMetadataKey: authToken,
	})
	if err := service.Forward(ctx, payload); err != nil {
		return err
	}
	fmt.Printf("push relay running profile=%s port=%s\n", *pushProfile, *relayPort)
	return nil
}
