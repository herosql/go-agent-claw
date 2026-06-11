package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/scene/registration"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := registration.RegisterApp(ctx, &registration.Options{
		OnQRCode: func(info *registration.QRCodeInfo) {
			fmt.Printf("open or scan this url: %s\n", info.URL)
			fmt.Printf("the link expires in %d seconds\n", info.ExpireIn)
		},
		OnStatusChange: func(info *registration.StatusChangeInfo) {
			// status: polling | slow_down | domain_switched
			fmt.Printf("registration status: %s", info.Status)
			if info.Interval > 0 {
				fmt.Printf(", next poll after %d seconds", info.Interval)
			}
			fmt.Println()
		},
	})
	if err != nil {
		var regErr *registration.RegisterAppError
		if errors.As(err, &regErr) {
			fmt.Printf("register app failed: code=%s, description=%s\n", regErr.Code, regErr.Description)
			return
		}
		panic(err)
	}

	fmt.Println("App ID:", result.ClientID)
	fmt.Println("App Secret:", result.ClientSecret)

	client := lark.NewClient(result.ClientID, result.ClientSecret)

	_ = client
}
