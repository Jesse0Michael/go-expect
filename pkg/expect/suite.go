package expect

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

// Connection is a named connection to a service under test.
type Connection interface {
	GetName() string
	Type() string
}

// Suite holds a collection of scenarios to run.
type Suite struct {
	scenarios   []*Scenario
	connections map[string]Connection
	defaultConn Connection
	log         *slog.Logger
}

// NewSuite creates an empty Suite.
func NewSuite() *Suite {
	return &Suite{
		connections: make(map[string]Connection),
		log:         slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

// WithLogger overrides the logger used when running scenarios.
func (s *Suite) WithLogger(log *slog.Logger) *Suite {
	s.log = log
	return s
}

// WithConnections registers one or more named connections.
// The first connection registered becomes the default for steps with no explicit connection.
func (s *Suite) WithConnections(conns ...Connection) *Suite {
	for _, c := range conns {
		s.connections[c.GetName()] = c
		if s.defaultConn == nil {
			s.defaultConn = c
		}
	}
	return s
}

// WithScenarios appends scenarios to the suite.
func (s *Suite) WithScenarios(scenarios ...*Scenario) *Suite {
	s.scenarios = append(s.scenarios, scenarios...)
	return s
}

// Run executes all scenarios. Each scenario gets its own fresh VarStore.
func (s *Suite) Run() error {
	var errs []error
	for _, sc := range s.scenarios {
		vars := make(VarStore)
		if err := sc.Run(s.log, s.defaultConn, s.connections, vars); err != nil {
			errs = append(errs, fmt.Errorf("scenario %q: %w", sc.Name, err))
		}
	}
	return errors.Join(errs...)
}
