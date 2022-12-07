package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	listNotificationOptions := &github.NotificationListOptions{}
	notifictations, response, err := client.Activity.ListNotifications(ctx, listNotificationOptions)
	if err != nil {
		panic(err)
	}

	fmt.Printf("response: %+v\n", response)
	for _, note := range notifictations {
		fmt.Printf("%s: %s\n", *note.Reason, *note.Subject.Title)
		fmt.Printf("  %s\n", *note.Repository.HTMLURL)
	}
}
