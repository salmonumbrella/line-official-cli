package cmd

import "fmt"

// FlagCheck represents a named boolean condition for flag validation.
type FlagCheck struct {
	Name string
	Set  bool
}

// requireExactlyOneFlag validates that exactly one of the given flags is set.
// Each FlagCheck contains a flag description (used in error messages) and whether it's set.
// Returns nil if exactly one flag is set, otherwise returns an error.
func requireExactlyOneFlag(checks []FlagCheck) error {
	var setFlags []string
	for _, c := range checks {
		if c.Set {
			setFlags = append(setFlags, c.Name)
		}
	}

	switch len(setFlags) {
	case 0:
		// Build list of flag names for error message
		names := make([]string, len(checks))
		for i, c := range checks {
			names[i] = c.Name
		}
		return fmt.Errorf("specify one of %v", names)
	case 1:
		return nil
	default:
		return fmt.Errorf("specify only one message type, got: %v", setFlags)
	}
}
