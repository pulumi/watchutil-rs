// Copyright 2016-2022, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/google/go-github/v47/github"
	"github.com/pulumi/watchutil-rs/tools/changelog/changelog"
	"github.com/spf13/cobra"
)

var inDir string
var version string
var format string
var useGitHub bool
var filterSinceCommit string
var filterOpenPrNumber int

// renderCmd represents the render command
var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: runRender,
}

func runRender(cmd *cobra.Command, args []string) error {
	configData, err := os.ReadFile(configFilename)
	if err != nil {
		return fmt.Errorf("unable to read config file at path %v: %w", configFilename, err)
	}
	var config changelog.Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return fmt.Errorf("unable to parse config file: %w", err)
	}
	inputFs := os.DirFS(filepath.Clean(inDir))
	fullChangelog := changelog.Changelog{}

	client := github.NewClient(nil)
	err = fs.WalkDir(inputFs, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		if d.IsDir() {
			return fs.SkipDir
		}
		if filepath.Ext(path) == ".yaml" {
			content, err := fs.ReadFile(inputFs, path)
			if err != nil {
				return err
			}

			var entry changelog.Changelog
			err = yaml.Unmarshal(content, &entry)
			if err != nil {
				return err
			}

			// entry.Changes[0].GitHubMeta.PullRequestNumbers = append(entry.Changes[0].GitHubMeta.PullRequestNumbers, 355)

			commitRange := "HEAD"
			if filterSinceCommit != "" {
				commitRange = filterSinceCommit + "..HEAD"
			}
			cmd := exec.Command("git", "rev-list", commitRange, filepath.Join(inDir, path))
			revlistOutput, err := cmd.Output()
			if err != nil {
				return nil
			}
			output := strings.TrimSpace(string(revlistOutput))
			revlist := strings.Split(output, "\n")
			if filterSinceCommit != "" && (len(output) == 0 || len(revlist) == 0) {
				return nil
			}
			for _, v := range revlist {
				prs, _, err := client.PullRequests.ListPullRequestsWithCommit(context.TODO(), "pulumi", "watchutil-rs", v, &github.PullRequestListOptions{})
				if err != nil {
					return err
				}
				for _, pr := range prs {
					// If we're previewing the changelog in a PR, there may be multiple open PRs containing the commit(s)
					// modifying this file. Only show open PRs related to the current commit.
					if filterOpenPrNumber != 0 && pr.GetState() == "open" && pr.GetNumber() != filterOpenPrNumber {
						continue
					}
					for _, c := range entry.Changes {
						c.GitHubMeta.PullRequestNumbers = append(c.GitHubMeta.PullRequestNumbers, pr.GetNumber())
					}
				}
			}

			fullChangelog = fullChangelog.Merge(entry)
		}

		return nil
	})
	if err != nil {
		return err
	}

	var date = time.Now().UTC().Format("2006-01-02")

	if format == "markdown" {
		buf, err := fullChangelog.Markdown(config, version, date)
		if err != nil {
			return err
		}
		fmt.Print(buf.String())
	} else if format == "conventional" {
		var buf strings.Builder
		for i, v := range fullChangelog.Changes {
			if i > 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(v.Conventional())
		}
		fmt.Print(buf.String())
	}

	return nil
}

func init() {
	rootCmd.AddCommand(renderCmd)

	renderCmd.Flags().SortFlags = false
	renderCmd.Flags().StringVarP(&configFilename, "config", "c", "changelog/config.yaml", "Config file")
	renderCmd.Flags().StringVarP(&format, "format", "f", "markdown", "Format")
	renderCmd.Flags().StringVarP(&inDir, "input", "i", "changelog/pending", "Input directory")
	renderCmd.Flags().StringVarP(&version, "version", "v", "", "Version")
	renderCmd.Flags().BoolVarP(&useGitHub, "github", "", true, "Use GitHub to resolve pull request numbers")
	renderCmd.Flags().StringVarP(&filterSinceCommit, "filter-since-commit", "", "", "Filter changes to render to those since a given commitish")
	renderCmd.Flags().IntVarP(&filterOpenPrNumber, "filter-open-pr-number", "", 0, "Filter PR number detection to a particular open PR number.")
}
