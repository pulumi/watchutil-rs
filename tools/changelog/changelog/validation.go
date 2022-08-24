package changelog

import (
	"fmt"
	"strings"
)

// ValidateType ensures a type string is in the permitted list and always returns a non-empty string if given a
// non-empty string to allow "--force" behavior.
func ValidateType(config Config, typ string) (string, error) {
	typ = strings.ToLower(strings.TrimSpace(typ))
	for _, v := range config.Types.Keys() {
		if typ == v {
			return typ, nil
		}
	}

	return typ, fmt.Errorf("unknown entry type %q", typ)
}

func PermittedTypesString(config Config) string {
	return strings.Join(config.Types.Keys(), ", ")
}

// var languages []string = []string{
// 	"dotnet",
// 	"go",
// 	"java",
// 	"nodejs",
// 	"python",
// 	"yaml",
// 	"docs",
// }

// // Each scope is a prefix, then an optional list of sub-scopes.
// //
// // Derived from historical changelog entries with small refinements.
// var PermittedScopes map[string][]string = map[string][]string{
// 	"auto":    languages,
// 	"backend": {"service", "filestate"},
// 	"cli": append([]string{
// 		"about",
// 		"config",
// 		"display",
// 		"engine",
// 		"import",
// 		"new",
// 		"plugin",
// 	}, languages...),
// 	"pkg": {"testing"},
// 	"sdk": languages,
// 	// Formerly known as multilang
// 	"components": languages,
// 	"engine":     nil,
// 	"build":      nil,
// 	"ci":         nil,
// 	"docs":       nil,
// 	"programgen": languages,
// 	"sdkgen":     languages,
// 	"protobuf":   nil,
// }

func PermittedScopesString(config Config, breakAfterItem bool) string {
	var buf string
	keys := config.Scopes.Keys()

	spacing := ", "
	if breakAfterItem {
		spacing = "\n"
	}

	for i, scope := range keys {
		if i != 0 {
			buf += spacing
		}

		buf += scope

		subs, _ := config.Scopes.Get(scope)
		if len(subs) > 0 {
			buf += "/" + strings.Join(subs, ",") + ""
		}
	}
	return buf
}
