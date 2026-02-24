package expect

import (
	"errors"
	"fmt"
	"log/slog"
)

// AfterFunc is a cleanup function run after all steps complete.
type AfterFunc func() error

// BeforeFunc is a setup function run before steps execute.
type BeforeFunc func() error

// Scenario is a named sequence of steps executed against one or more connections.
type Scenario struct {
	Name   string
	steps  []Step
	before []BeforeFunc
	after  []AfterFunc
}

// NewScenario creates a new Scenario with the given name.
func NewScenario(name string) *Scenario {
	return &Scenario{Name: name}
}

// AddStep appends a step to the scenario.
func (s *Scenario) AddStep(b *StepBuilder) *Scenario {
	s.steps = append(s.steps, b.Build())
	return s
}

// Before registers a function to run before the scenario's steps.
func (s *Scenario) Before(fn BeforeFunc) *Scenario {
	s.before = append(s.before, fn)
	return s
}

// After registers a cleanup function to always run after the scenario.
func (s *Scenario) After(fn AfterFunc) *Scenario {
	s.after = append(s.after, fn)
	return s
}

// Run executes all steps sequentially, then runs all after-funcs.
// after-funcs always execute regardless of before or step failures.
func (s *Scenario) Run(log *slog.Logger, defaultConn Connection, connections map[string]Connection, vars VarStore) error {
	log = log.With("scenario", s.Name)
	log.Info("starting scenario")

	var errs []error

	for _, fn := range s.before {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		for i, step := range s.steps {
			conn := defaultConn
			if step.Connection != "" {
				if c, ok := connections[step.Connection]; ok {
					conn = c
				}
			}

			label := stepLabel(i, step)
			log.Info("step", "step", label)

			if err := step.Run(conn, vars); err != nil {
				log.Error("step failed", "step", label, "error", err)
				errs = append(errs, fmt.Errorf("step %s: %w", label, err))
			} else {
				log.Info("step passed", "step", label)
			}
		}
	}

	for _, fn := range s.after {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		log.Error("scenario failed", "errors", len(errs))
	} else {
		log.Info("scenario passed")
	}
	return errors.Join(errs...)
}

func stepLabel(i int, s Step) string {
	if s.Request == nil {
		return fmt.Sprintf("[%d] (no request)", i+1)
	}
	switch req := s.Request.(type) {
	case *HTTPRequest:
		return fmt.Sprintf("[%d] %s %s", i+1, req.Method, req.Path)
	case *GRPCRequest:
		return fmt.Sprintf("[%d] grpc %s", i+1, req.FullMethod)
	default:
		return fmt.Sprintf("[%d]", i+1)
	}
}
