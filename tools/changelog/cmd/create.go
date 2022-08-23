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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/gosimple/slug"
	"github.com/pulumi/watchutil-rs/tools/changelog/changelog"
	"github.com/spf13/cobra"

	"github.com/erikgeiser/promptkit/selection"
)

func newCreateCmd() *cobra.Command {
	var typ string
	var scope string
	var desc string
	var title string
	var force bool
	var issues []int
	var prs []int
	var outDir string
	var configFilename string

	// createCmd represents the create command
	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a changelog entry",
		Long: `Create a changelog entry. If not provided by arguments, prompts for the type, scope, ` +
			`and description of the change and the title of the entry file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configData, err := os.ReadFile(configFilename)
			if err != nil {
				return fmt.Errorf("unable to read config file at path %v: %w", configFilename, err)
			}
			var config changelog.Config
			err = yaml.Unmarshal(configData, &config)
			if err != nil {
				return fmt.Errorf("unable to parse config file: %w", err)
			}

			err = os.MkdirAll(outDir, 0700)
			if err != nil {
				return err
			}

			// Get type from argument or prompt:
			if typ == "" {
				typ, err = typePrompt(config, err)
			} else {
				typ, err = changelog.ValidateType(config, typ)
			}
			if err != nil && !force {
				return err
			}

			// Get scopes from argument or prompt:
			var subs []string
			if scope == "" {
				scope, subs, err = scopePrompt(config)
			} else {
				scope, subs, err = changelog.ParseScope(config, scope, force)
			}
			if err != nil && !force {
				return err
			}

			// Get desc from argument or prompt:
			if len(desc) == 0 {
				for {
					desc, err = textPrompt("Description of change?", "Fixes bug, adds feature, updates dependencies, etc.", desc)
					if err != nil {
						return err
					}

					confirmed, err := confirmationPrompt()
					if err != nil {
						return err
					}
					if confirmed {
						clearBack(1)
						break
					} else {
						clearBack(2)
						continue
					}
				}
			}

			change := changelog.Entry{
				Type: typ,
				Scope: changelog.Scope{
					Primary:   scope,
					SubScopes: subs,
				},
				Description: desc,
				GitHubMeta: changelog.GitHubMeta{
					PullRequestNumbers: prs,
				},
			}
			cl := changelog.Changelog{
				Entries: []*changelog.Entry{&change},
			}

			entry, err := yaml.Marshal(cl)
			if err != nil {
				return err
			}

			var date = time.Now().UTC().Format("20060102")

			if len(title) == 0 {
				for {
					title, err = textPrompt("Title of change entry?", "issue-1, convert-stack-references, etc.", title)
					if err != nil {
						return err
					}

					title = slug.Make(title)
					fmt.Printf("Title will be: %v\n", renderTitle(date, change, title))

					confirmed, err := confirmationPrompt()
					if err != nil {
						return err
					}
					if confirmed {
						clearBack(1)
						break
					} else {
						clearBack(3)
						continue
					}
				}
			}

			sort.Strings(subs)
			sort.Ints(issues)
			sort.Ints(prs)

			filename := renderTitle(date, change, title)
			filename = filepath.Join(outDir, filename)
			fmt.Printf("Creating change entry: %v\n", filename)

			return os.WriteFile(filename, entry, 0600)
		},
	}

	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVarP(&configFilename, "config", "c", "changelog/config.yaml", "Config file")
	createCmd.Flags().StringVarP(&outDir, "out", "o", "changelog/pending", "Output directory")

	typeUsage := "Type of entry, as defined by config"
	createCmd.Flags().StringVarP(&typ, "type", "t", "", typeUsage)

	scopeUsage := "Scope of entry, as defined by config, with optional\n" +
		"comma-delimited subscopes."
	createCmd.Flags().StringVarP(&scope, "scope", "s", "", scopeUsage)

	createCmd.Flags().StringVarP(&desc, "description", "d", "", "Description of change")
	createCmd.Flags().StringVarP(&title, "title", "", "", "Title of entry to create")
	createCmd.Flags().IntSliceVarP(&issues, "issue", "i", []int{}, "Issues addressed by change.")
	createCmd.Flags().IntSliceVarP(&prs, "pull-request", "p", []int{}, "Pull request implementing change.")
	createCmd.Flags().BoolVarP(&force, "force", "f", false, "Allows unknown types and scopes to be specified.")

	return createCmd
}

func renderTitle(date string, change changelog.Entry, title string) string {
	slug.Lowercase = false
	scope := slug.Make(change.Scope.String())
	scopePrefix := ""
	if scope != "" {
		scopePrefix = "--"
	}
	filename := fmt.Sprintf("%v%v%v--%v.yaml",
		slug.Make(date), scopePrefix, scope, slug.Make(title))
	return filename
}

func typePrompt(config changelog.Config, err error) (string, error) {
	for {
		sp := selection.New("Type of change?", config.Types.Keys())
		sp.Filter = nil

		typ, err := sp.RunPrompt()
		if err != nil {
			return "", err
		}

		confirmed, err := confirmationPrompt()
		if err != nil {
			return "", err
		}
		if confirmed {
			clearBack(1)
			return typ, nil
		} else {
			clearBack(2)
			continue
		}
	}
}

func init() {
	rootCmd.AddCommand(newCreateCmd())
}
