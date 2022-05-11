package ghutil

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/exp/slices"
	"golang.org/x/oauth2"
)

func GetClientFromToken(ctx context.Context, token string) *github.Client {
	// Authenticate with static token
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Create a new GitHub client
	return github.NewClient(tc)
}

const (
	ReasonAssigned        = "assigned"
	ReasonComment         = "comment"
	ReasonMention         = "mention"
	ReasonReviewRequested = "review_requested"
	ReasonTeamMention     = "team_mention"
	ReasonAuthor          = "author"
)

var (
	// From lowest to highest priority
	ReasonPriority = []string{ReasonAssigned, ReasonAuthor, ReasonReviewRequested, ReasonTeamMention, ReasonMention, ReasonComment}
)

func GetUnreadPullRequests(ctx context.Context, client *github.Client) ([]*github.Notification, error) {
	// list all notifications
	notificationList, _, err := client.Activity.ListNotifications(ctx, nil)
	if err != nil {
		return nil, err
	}

	// filter to only unread pr notifications
	notifications := make([]*github.Notification, 0)
	for _, notification := range notificationList {
		if *notification.Unread && *notification.Subject.Type == "PullRequest" {
			notifications = append(notifications, notification)
		}
	}

	sort.Slice(notifications, func(i, j int) bool {
		if *notifications[i].Repository.FullName != *notifications[j].Repository.FullName {
			return *notifications[i].Repository.FullName < *notifications[j].Repository.FullName
		}
		return slices.Index(ReasonPriority, *notifications[i].Reason) > slices.Index(ReasonPriority, *notifications[j].Reason)
	})

	return notifications, nil
}

func GetPullRequestURL(notification *github.Notification) string {
	return fmt.Sprintf("https://github.com/%s/pull/%d?notification_referrer_id=%s", *notification.Repository.FullName, GetPullRequestID(notification), *notification.ID)
}

func GetPullRequestID(notification *github.Notification) int {
	parts := strings.Split(*notification.Subject.URL, "/")
	val, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return -1
	}
	return val
}
