package achievements

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

func Test_versionInWindow(t *testing.T) {
	for name, tc := range map[string]struct {
		source   *version.Version
		from     *version.Version
		to       *version.Version
		expected bool
	}{
		"less": {
			source:   version.Must(version.NewSemver("0.9.99")),
			from:     version.Must(version.NewSemver("1.0.0")),
			to:       version.Must(version.NewSemver("1.1.0")),
			expected: false,
		},
		"more": {
			source:   version.Must(version.NewSemver("1.1.11")),
			from:     version.Must(version.NewSemver("1.0.0")),
			to:       version.Must(version.NewSemver("1.1.0")),
			expected: false,
		},
		"in": {
			source:   version.Must(version.NewSemver("1.0.11")),
			from:     version.Must(version.NewSemver("1.0.0")),
			to:       version.Must(version.NewSemver("1.1.0")),
			expected: true,
		},
		"equal from": {
			source:   version.Must(version.NewSemver("1.0.0")),
			from:     version.Must(version.NewSemver("1.0.0")),
			to:       version.Must(version.NewSemver("1.1.0")),
			expected: true,
		},
		"equal to": {
			source:   version.Must(version.NewSemver("1.1.0")),
			from:     version.Must(version.NewSemver("1.0.0")),
			to:       version.Must(version.NewSemver("1.1.0")),
			expected: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			actual := versionInWindow(tc.source, tc.from, tc.to)
			require.Equal(t, tc.expected, actual)
		})
	}
}
