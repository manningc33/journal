package run

import "context"

// Call records a single invocation made against a Fake runner.
type Call struct {
	Kind string // "interactive" or "capture"
	Name string
	Args []string
}

// Fake is a Runner that records calls instead of executing them, for tests.
type Fake struct {
	Calls []Call

	// NotFound lists binary names that LookPath should report as missing.
	NotFound map[string]bool
	// CaptureOut and CaptureErr are returned from Capture.
	CaptureOut []byte
	CaptureErr error
	// InteractiveErr is returned from Interactive.
	InteractiveErr error
}

// Interactive implements Runner.
func (f *Fake) Interactive(_ context.Context, name string, args ...string) error {
	f.Calls = append(f.Calls, Call{Kind: "interactive", Name: name, Args: args})
	return f.InteractiveErr
}

// Capture implements Runner.
func (f *Fake) Capture(_ context.Context, name string, args ...string) ([]byte, error) {
	f.Calls = append(f.Calls, Call{Kind: "capture", Name: name, Args: args})
	return f.CaptureOut, f.CaptureErr
}

// LookPath implements Runner.
func (f *Fake) LookPath(name string) (string, error) {
	if f.NotFound[name] {
		return "", &lookupError{name}
	}
	return "/usr/bin/" + name, nil
}

type lookupError struct{ name string }

func (e *lookupError) Error() string { return "executable not found: " + e.name }
