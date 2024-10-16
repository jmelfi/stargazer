package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetAt   time.Time `json:"reset_at"`
}

// Star represents a starred GitHub repository with its details.
type Star struct {
	Url           string    // Repository URL
	Name          string    // Repository name
	NameWithOwner string    // Repository name with owner (e.g., "owner/repo")
	Description   string    // Repository description
	License       string    // Repository license
	LicenseUrl    string    // URL to the license
	Stars         int       // Number of stars
	Archived      bool      // Whether the repository is archived
	StarredAt     time.Time // When the repository was starred by the user
}

var query struct {
	RateLimit struct {
		Limit     int
		Remaining int
		ResetAt   time.Time
	}
	User struct {
		StarredRepositories struct {
			IsOverLimit bool
			TotalCount  int
			Edges       []struct {
				StarredAt time.Time
				Node      struct {
					Description string
					Languages   struct {
						Edges []struct {
							Node struct {
								Name string
							}
						}
					} `graphql:"languages(first: $lc, orderBy: {field: SIZE, direction: DESC})"`
					LicenseInfo struct {
						Name     string
						Nickname string
						Url      string
					}
					IsArchived     bool
					IsPrivate      bool
					Name           string
					NameWithOwner  string
					StargazerCount int
					Url            string
				}
			}
			PageInfo struct {
				EndCursor   string
				HasNextPage bool
			}
		} `graphql:"starredRepositories(first: $count, orderBy: {field: STARRED_AT, direction: DESC}, after: $cursor)"`
	} `graphql:"user(login: $login)"`
}

// FetchStarsFunc is the function type for fetching stars
type FetchStarsFunc func(user string, token string, rateLimit int) (map[string][]Star, int, error)

// DefaultFetchStars is the default implementation of FetchStarsFunc
var DefaultFetchStars FetchStarsFunc = func(user string, token string, rateLimit int) (map[string][]Star, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, src)

	client := githubv4.NewClient(httpClient)

	vars := map[string]interface{}{
		"login":  githubv4.String(user),
		"lc":     githubv4.Int(1),
		"count":  githubv4.Int(50),
		"cursor": githubv4.String(""),
	}

	stars := make(map[string][]Star)
	total := 0

	rateLimiter := rate.NewLimiter(rate.Every(time.Second/time.Duration(rateLimit)), 1)

	rateLimitInfo, err := loadRateLimitInfo()
	if err != nil {
		logger.WithError(err).Warn("Failed to load rate limit info, using default")
	} else {
		logger.WithFields(logrus.Fields{
			"remaining": rateLimitInfo.Remaining,
			"reset_at":  rateLimitInfo.ResetAt,
		}).Debug("Loaded GitHub API rate limit info")
	}

	for {
		if err := rateLimiter.Wait(ctx); err != nil {
			logger.WithError(err).Error("Rate limit exceeded")
			return stars, total, err
		}

		err = client.Query(ctx, &query, vars)
		if err != nil {
			if isRateLimitError(err) {
				logger.WithError(err).Warn("Rate limit reached, waiting before retry")
				time.Sleep(time.Until(query.RateLimit.ResetAt))
				continue
			}
			logger.WithError(err).Error("Failed to query GitHub API")
			return stars, total, err
		}

		rateLimitInfo = RateLimitInfo{
			Limit:     query.RateLimit.Limit,
			Remaining: query.RateLimit.Remaining,
			ResetAt:   query.RateLimit.ResetAt,
		}
		if err := saveRateLimitInfo(rateLimitInfo); err != nil {
			logger.WithError(err).Warn("Failed to save rate limit info")
		}

		logger.WithFields(logrus.Fields{
			"remaining": rateLimitInfo.Remaining,
			"reset_at":  rateLimitInfo.ResetAt,
		}).Debug("GitHub API rate limit status")

		for _, e := range query.User.StarredRepositories.Edges {
			if e.Node.IsPrivate || isIgnored(e.Node.NameWithOwner) {
				continue
			}

			total++
			lng := determineLanguage(e.Node.Languages.Edges)
			if _, ok := stars[lng]; !ok {
				stars[lng] = make([]Star, 0)
			}

			lic := determineLicense(e.Node.LicenseInfo)

			stars[lng] = append(stars[lng], Star{
				Url:           e.Node.Url,
				Name:          e.Node.Name,
				NameWithOwner: e.Node.NameWithOwner,
				Description:   e.Node.Description,
				License:       lic,
				LicenseUrl:    e.Node.LicenseInfo.Url,
				Stars:         e.Node.StargazerCount,
				Archived:      e.Node.IsArchived,
				StarredAt:     e.StarredAt,
			})
		}

		if !query.User.StarredRepositories.PageInfo.HasNextPage {
			break
		}
		vars["cursor"] = githubv4.String(query.User.StarredRepositories.PageInfo.EndCursor)
	}

	logger.WithField("total_stars", total).Info("Successfully fetched starred repositories")
	return stars, total, nil
}

func loadRateLimitInfo() (RateLimitInfo, error) {
	data, err := os.ReadFile("rate_limit_info.json")
	if err != nil {
		return RateLimitInfo{}, err
	}

	var info RateLimitInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return RateLimitInfo{}, err
	}

	return info, nil
}

func saveRateLimitInfo(info RateLimitInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return os.WriteFile("rate_limit_info.json", data, 0644)
}

func isRateLimitError(err error) bool {
	return strings.Contains(err.Error(), "API rate limit exceeded")
}

func determineLanguage(languages []struct{ Node struct{ Name string } }) string {
	if len(languages) > 0 {
		lang := languages[0].Node.Name
		logger.WithField("language", lang).Debug("Determined repository language")
		return lang
	}
	logger.Debug("No language determined for repository")
	return "Unknown"
}

func determineLicense(licenseInfo struct {
	Name     string
	Nickname string
	Url      string
}) string {
	var license string
	if licenseInfo.Nickname != "" {
		license = licenseInfo.Nickname
	} else if licenseInfo.Name != "" && strings.ToLower(licenseInfo.Name) != "other" {
		license = licenseInfo.Name
	}

	if license != "" {
		logger.WithField("license", license).Debug("Determined repository license")
	} else {
		logger.Debug("No license determined for repository")
	}

	return license
}
