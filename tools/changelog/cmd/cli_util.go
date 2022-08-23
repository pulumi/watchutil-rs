package cmd

import (
	"fmt"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/muesli/termenv"
	"github.com/pulumi/watchutil-rs/tools/changelog/changelog"
)

func textPrompt(textPrompt string, textPlaceholder string, textValue string) (string, error) {

	input := textinput.New(textPrompt)
	input.Placeholder = textPlaceholder
	input.InitialValue = textValue

	textValue, err := input.RunPrompt()
	if err != nil {
		return "", err
	}

	return textValue, nil
}

func confirmationPrompt() (bool, error) {
	input := textinput.New("Confirm [Y/n]?")
	input.Placeholder = ""
	input.Validate = func(answer string) error {
		answer = strings.ToLower(answer)

		if len(answer) == 0 || strings.HasPrefix("yes", answer) || strings.HasPrefix("no", answer) {
			return nil
		}

		return fmt.Errorf("expected yes or no")
	}

	confirmation, err := input.RunPrompt()
	if err != nil {
		return false, err
	}
	confirmation = strings.ToLower(confirmation)
	if len(confirmation) == 0 || strings.HasPrefix("yes", confirmation) {
		return true, nil
	} else if strings.HasPrefix("no", confirmation) {
		return false, nil
	}

	panic("TODO, unreachable")
}

func scopePrompt(config changelog.Config) (string, []string, error) {
	var err error
	var scope string
	var subs []string
	scopes := config.Scopes.Keys()

	if len(scopes) == 0 {
		return "", nil, nil
	}

	for {
		sp := selection.New("Scope of change?", scopes)
		scope, err = sp.RunPrompt()
		if err != nil {
			return "", nil, err
		}

		confirmed, err := confirmationPrompt()
		if err != nil {
			return "", nil, err
		}
		if confirmed {
			clearBack(1)
			break
		} else {
			clearBack(2)
			continue
		}
	}

	selectedSubs := make(map[string]any)
	for {
		subs = []string{}
		subScopes := []string{}
		if subScopeList, ok := config.Scopes.Get(scope); ok {
			subScopes = append([]string{}, subScopeList...)
		}

		for i, v := range subScopes {
			if _, has := selectedSubs[v]; has {
				subs = append(subs, v)
				subScopes[i] = "[X] " + v
			} else {
				subScopes[i] = "[ ] " + v
			}
		}
		subScopes = append(subScopes, "    Done")
		subScopes = append(subScopes, "    Reset")

		scopeText := changelog.Scope{Primary: scope, SubScopes: subs}.String()
		sp := selection.New(fmt.Sprintf("Select sub-scopes? Current scope: %v", renderCurrentChoice(scopeText)), subScopes)
		sp.FinalChoiceStyle = func(c *selection.Choice[string]) string {
			return ""
		}
		choiceRaw, err := sp.RunPrompt()
		if err != nil {
			return "", nil, err
		}
		choice := choiceRaw[4:]

		if choice == "Done" {
			break
		} else if choice == "Reset" {
			selectedSubs = make(map[string]any)
		} else {
			if _, has := selectedSubs[choice]; has {
				delete(selectedSubs, choice)
			} else {
				selectedSubs[choice] = struct{}{}
			}
		}
		clearBack(1)
	}
	return scope, subs, nil
}

func renderCurrentChoice(choice string) string {
	return termenv.String(choice).Foreground(termenv.ANSI256Color(32)).String()
}

func clearLine() {
	fmt.Printf("\033[2K")
}
func cursorUp(lines int) {
	fmt.Printf("\033[%vA", lines)
}

func clearBack(lines int) {
	clearLine()
	for i := 0; i < lines; i++ {
		cursorUp(1)
		clearLine()
	}
}
