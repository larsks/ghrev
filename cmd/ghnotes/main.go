package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/google/go-github/v48/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Fatalf("ERROR: %v", err)
		}
	}()

	if cfgPath, err := xdg.ConfigFile("ghnotes/env"); err == nil {
		if err := godotenv.Load(cfgPath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Printf("WARNING: unable to load .env file: %v", err)
			}
		}
	}

	if err := godotenv.Load(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("WARNING: unable to load .env file: %v", err)
		}
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	listNotificationOptions := &github.NotificationListOptions{}
	notifictations, _, err := client.Activity.ListNotifications(ctx, listNotificationOptions)
	if err != nil {
		panic(fmt.Errorf("failed to authenticate to github: %w", err))
	}

	for _, note := range notifictations {
		var tag string
		switch note.GetReason() {
		case "review_requested":
			tag = "RV"
		case "ci_activity":
			tag = "CI"
		case "mentioned":
			tag = "M "
		case "assigned":
			tag = "A "
		case "author":
			tag = "AU"
		case "subscribed":
			tag = "S "
		default:
			tag = "  "
		}

		var html_url string
		subject := note.GetSubject()
		if url := subject.GetURL(); url != "" {
			switch subject.GetType() {
			case "PullRequest":
				urlparts := strings.Split(subject.GetURL(), "/")
				pullNumber, err := strconv.Atoi(urlparts[7])
				if err != nil {
					panic(err)
				}
				pull, _, err := client.PullRequests.Get(context.Background(), urlparts[4], urlparts[5], pullNumber)
				if err != nil {
					log.Printf("WARNING: failed to fetch pull request: %v", err)
					break
				}
				if pull.GetState() == "closed" {
					continue
				}
				html_url = pull.GetHTMLURL()
			case "Issue":
				urlparts := strings.Split(subject.GetURL(), "/")
				issueNumber, err := strconv.Atoi(urlparts[7])
				if err != nil {
					panic(err)
				}
				issue, _, err := client.Issues.Get(context.Background(), urlparts[4], urlparts[5], issueNumber)
				if err != nil {
					log.Printf("WARNING: failed to fetch issue: %v", err)
					break
				}
				if issue.GetState() == "closed" {
					continue
				}
				html_url = issue.GetHTMLURL()
			}
		}

		if html_url == "" {
			html_url = note.Repository.GetHTMLURL()
		}

		fmt.Printf("[%2s] %s\n", tag, note.Repository.GetFullName())
		fmt.Printf("    %s\n", note.GetSubject().GetTitle())
		fmt.Printf("    %s\n", html_url)
	}
}
