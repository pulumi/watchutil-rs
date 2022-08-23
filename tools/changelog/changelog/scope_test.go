package changelog

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func TestComplexScopeRoundtrip(t *testing.T) {
	t.Parallel()

	keys := make([]string, 0, len(PermittedScopes))
	for k := range PermittedScopes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := 0; i < 1; i++ {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			for _, chosenScope := range keys { //nolint:paralleltest // overhead from parallelization
				permittedSubs := append([]string{}, PermittedScopes[chosenScope]...)

				var chosenSubs []string
				if len(permittedSubs) < 2 {
					continue
				}

				// Choose 2..n subscopes
				subscopesToChoose := rand.Intn(len(permittedSubs)) //nolint: gosec
				for i := 0; i < subscopesToChoose; i++ {
					idx := rand.Intn(len(permittedSubs)) //nolint: gosec
					sub := permittedSubs[idx]
					chosenSubs = append(chosenSubs, sub)
					permittedSubs = removeRandom(permittedSubs, idx)
				}

				var scopeInput string
				if len(chosenSubs) == 0 {
					scopeInput = chosenScope
				} else {
					scopeInput = chosenScope + "/" + strings.Join(chosenSubs, ",")
				}

				t.Run(scopeInput, func(t *testing.T) {
					var scope Scope
					err := yaml.Unmarshal([]byte(scopeInput), &scope)
					assert.NoError(t, err)
					if err != nil {
						return
					}
					assert.Equalf(t, chosenScope, scope.Primary, "parsed from input %v", scopeInput)
					assert.ElementsMatch(t, chosenSubs, scope.SubScopes)
					marshaled, err := yaml.Marshal(scope)
					assert.NoError(t, err)
					assert.Equal(t, scopeInput, strings.TrimSpace(string(marshaled)))
				})
			}
		})
	}
}
