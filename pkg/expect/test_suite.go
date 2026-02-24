package expect

import (
	"log/slog"
	"testing"
)

// TestSuite wraps Suite for use with Go's testing package.
type TestSuite struct {
	suite *Suite
}

// NewTestSuite creates a TestSuite backed by the given Suite.
func NewTestSuite(suite *Suite) *TestSuite {
	return &TestSuite{suite: suite}
}

// Run executes all scenarios, routing log output through t and reporting
// failures via t.Errorf so all scenarios always run (non-fatal).
func (s *TestSuite) Run(t *testing.T) {
	t.Helper()
	s.suite.WithLogger(slog.New(slog.NewTextHandler(t.Output(), nil)))
	if err := s.suite.Run(); err != nil {
		t.Errorf("suite failures:\n%v", err)
	}
}
