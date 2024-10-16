// Package main implements the Stargazer application, which creates lists of starred GitHub repositories.
package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd     *cobra.Command
	generateCmd *cobra.Command
)

const (
	appName = "stargazer"
	appDesc = "Creates awesome lists of your starred GitHub repositories"

	defaultOutput      = "README.md"
	defaultFormat      = "list"
	defaultWithToc     = true
	defaultWithStars   = true
	defaultWithLicense = true
	defaultWithBtt     = false

	envUser   = "GITHUB_USER"
	envToken  = "GITHUB_TOKEN"
	envOutput = "OUTPUT_FILE"
	envFormat = "OUTPUT_FORMAT"
	envIgnore = "IGNORE_REPOS"

	envToc     = "WITH_TOC"
	envStars   = "WITH_STARS"
	envLicense = "WITH_LICENSE"
	envBttLink = "WITH_BACK_TO_TOP"
)

var (
	version = ""
	ignored []string
	env     map[string]string
)

// main is the entry point of the Stargazer application.
func main() {
	if version == "" {
		version = "dev"
	}

	initConfig()

	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Fatal("Failed to execute root command")
	}
}

func initConfig() {
	viper.SetConfigName("stargazer")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logger.Info("Using config file:", viper.ConfigFileUsed())
	}
}

// parseConfig processes command-line flags and config file to build the application configuration.
func init() {
	rootCmd = &cobra.Command{
		Use:   appName,
		Short: appDesc,
		Long:  appDesc,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate the starred repositories list",
		Run:   runGenerate,
	}

	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringP("output-file", "o", defaultOutput, "the file to create")
	generateCmd.Flags().StringP("output-format", "f", defaultFormat, "the format of the output ["+strings.Join(availableFormats, ", ")+"]")
	generateCmd.Flags().StringP("github-user", "u", "", "github user name")
	generateCmd.Flags().String("github-token", "", "github access token")
	generateCmd.Flags().Int("rate-limit", 5, "number of API requests per second")
	generateCmd.Flags().StringSliceP("ignore", "i", []string{}, "repositories to ignore (flag can be specified multiple times)")
	generateCmd.Flags().BoolP("test", "t", false, "just put out some test data")
	generateCmd.Flags().Bool("with-toc", true, "print table of contents")
	generateCmd.Flags().Bool("with-stars", true, "print starcount of repositories")
	generateCmd.Flags().Bool("with-license", true, "print license of repositories")
	generateCmd.Flags().Bool("with-back-to-top", false, "generate 'back to top' links for each language")

	viper.BindPFlags(generateCmd.Flags())
}

func runGenerate(cmd *cobra.Command, args []string) {
	config := &Config{
		OutputFile:    viper.GetString("output-file"),
		OutputFormat:  viper.GetString("output-format"),
		GithubUser:    viper.GetString("github-user"),
		GithubToken:   viper.GetString("github-token"),
		IgnoreRepos:   viper.GetStringSlice("ignore"),
		Test:          viper.GetBool("test"),
		WithTOC:       viper.GetBool("with-toc"),
		WithStars:     viper.GetBool("with-stars"),
		WithLicense:   viper.GetBool("with-license"),
		WithBackToTop: viper.GetBool("with-back-to-top"),
		RateLimit:     viper.GetInt("rate-limit"),
	}

	if config.GithubToken == "" && !config.Test {
		logger.Fatal("GitHub token is required. Please provide a valid token.")
	}

	if err := initTemplate(config.OutputFormat); err != nil {
		logger.WithError(err).Fatal("Failed to initialize template")
	}

	stars, total, err := fetchAndProcessStars(config)
	if err != nil {
		logger.WithError(err).Fatal("Failed to fetch and process stars")
	}

	err = writeList(config.OutputFile, stars, total, config.WithTOC, config.WithLicense, config.WithStars, config.WithBackToTop)
	if err != nil {
		logger.WithError(err).Fatal("Failed to write list")
	}

	logger.WithField("total_repositories", total).Info("Successfully generated starred repositories list")
}

// fetchAndProcessStars retrieves and processes starred repositories based on the provided configuration.
func fetchAndProcessStars(config *Config) (map[string][]Star, int, error) {
	var stars map[string][]Star
	var total int
	var err error

	if config.Test {
		stars, total = testStars()
	} else {
		if stars, total, err = DefaultFetchStars(config.GithubUser, config.GithubToken, config.RateLimit); err != nil {
			return nil, 0, fmt.Errorf("failed to fetch stars: %v", err)
		}
	}

	for k, v := range stars {
		sort.Slice(v, func(i, j int) bool {
			return strings.ToLower(v[i].NameWithOwner) < strings.ToLower(v[j].NameWithOwner)
		})
		stars[k] = v
	}

	return stars, total, nil
}

// isIgnored checks if a repository name is in the ignored list.
func isIgnored(name string) bool {
	if len(ignored) == 0 {
		return false
	}
	for _, i := range ignored {
		if strings.ToLower(i) == strings.ToLower(name) {
			return true
		}
	}
	return false
}

// testStars generates test data for starred repositories.
func testStars() (stars map[string][]Star, total int) {
	stars = make(map[string][]Star)
	stars["go"] = make([]Star, 1)
	s := Star{
		Url:           "https://github.com/jmelfi/stargazer",
		Name:          "stargazer",
		NameWithOwner: "jmelfi/stargazer",
		Description:   "Creates awesome lists of your starred repositories",
		License:       "MIT License",
		Stars:         1,
		Archived:      false,
		StarredAt:     time.Now(),
	}
	if !isIgnored(s.NameWithOwner) {
		stars["go"][0] = s
	}
	stars["markdown"] = make([]Star, 1)
	s = Star{
		Url:           "https://github.com/jmelfi/stars",
		Name:          "stars",
		NameWithOwner: "jmelfi/stars",
		Description:   "A list of awesome repositories I starred",
		License:       "MIT License",
		Stars:         1,
		Archived:      false,
		StarredAt:     time.Now(),
	}
	if !isIgnored(s.NameWithOwner) {
		stars["markdown"][0] = s
	}

	stars["C#"] = make([]Star, 0)
	stars["C++"] = make([]Star, 0)

	stars["C#"] = append(stars["C#"], Star{
		Url:           "https://github.com/jmelfi/test",
		Name:          "test",
		NameWithOwner: "jmelfi/test",
		Description:   "",
		License:       "MIT License",
		Stars:         1,
		StarredAt:     time.Now(),
	})
	stars["C++"] = append(stars["C++"], Star{
		Url:           "https://github.com/jmelfi/test_2",
		Name:          "test_2",
		NameWithOwner: "rverst/test_2",
		Description:   "Some description",
		License:       "",
		Stars:         1,
		StarredAt:     time.Now(),
	})

	stars["C#"] = append(stars["C#"], Star{
		Url:           "https://github.com/jmelfi/test_3",
		Name:          "test_3",
		NameWithOwner: "jmelfi/test_3",
		Description:   "",
		License:       "",
		Stars:         1,
		StarredAt:     time.Now(),
	})

	total = 4
	return
}

// getEnv retrieves environment variables with fallback to .env file and default values.
func getEnv(key, defVal string) string {
	val := os.Getenv(key)
	if v, ok := env[key]; ok {
		val = v
	}
	if val == "" {
		return defVal
	}
	return val
}

// parseEnvFile reads and parses a .env file.
func parseEnvFile(file string) map[string]string {
	env := make(map[string]string)
	f, err := os.Open(file)
	if err != nil {
		logger.WithError(err).Warn("Could not open .env file")
		return env
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if s.Err() != nil {
			logger.WithError(s.Err()).Warn("Error scanning .env file")
			continue
		}
		l := strings.Trim(s.Text(), " ")
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		s := strings.SplitN(l, "=", 2)
		if len(s) == 2 {
			env[strings.TrimRight(s[0], " ")] = strings.TrimLeft(s[1], " ")
		}
	}
	return env
}

// exists checks if a file exists and is not a directory.
func exists(file string) bool {
	fi, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	return !fi.IsDir()
}
