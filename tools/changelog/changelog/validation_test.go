package changelog

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyScope(t *testing.T) {
	t.Parallel()

	scope, subs, err := ParseScope("", false)
	assert.NoError(t, err)
	assert.Equal(t, "", scope)
	assert.Len(t, subs, 0)
}

func TestSimpleScopes(t *testing.T) {
	t.Parallel()

	for k := range PermittedScopes { //nolint:paralleltest // overhead from parallelization
		t.Run(k, func(t *testing.T) {
			scope, subs, err := ParseScope(k, false)
			assert.NoError(t, err)
			assert.Equal(t, k, scope)
			assert.ElementsMatch(t, subs, []string{})
		})
	}
}

func TestSimpleSubs(t *testing.T) {
	t.Parallel()

	for k, subs := range PermittedScopes { //nolint:paralleltest // overhead from parallelization
		for _, sub := range subs {
			t.Run(k+"/"+sub, func(t *testing.T) {
				scope, subs, err := ParseScope(k+"/"+sub, false)
				assert.NoError(t, err)
				assert.Equal(t, k, scope)
				assert.ElementsMatch(t, subs, []string{sub})
			})
		}
	}
}

func TestComplexScopes(t *testing.T) {
	t.Parallel()

	keys := make([]string, 0, len(PermittedScopes))
	for k := range PermittedScopes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := 0; i < 20; i++ {
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
				} else /* if len(chosenSubs) >= 1 */ {
					// Randomly add some whitespace between items:
					sep := "," + strings.Repeat(" ", rand.Intn(2)) //nolint: gosec
					scopeInput = chosenScope + "/" + strings.Join(chosenSubs, sep) + ""
				}

				t.Run(scopeInput, func(t *testing.T) {
					scope, subs, err := ParseScope(scopeInput, false)
					assert.NoError(t, err)
					assert.Equal(t, chosenScope, scope)
					assert.ElementsMatch(t, subs, chosenSubs)
				})
			}
		})
	}
}

func removeRandom(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
