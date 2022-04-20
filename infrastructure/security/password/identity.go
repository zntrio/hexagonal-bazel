package password

// Identity is used for testing purpose.
func Identity() Strategy {
	return &identityStrategy{}
}

// -----------------------------------------------------------------------------

type identityStrategy struct {
}

func (s *identityStrategy) Hash(input string) (string, error) {
	return input, nil
}

func (s *identityStrategy) Verify(given, secret string) error {
	return nil
}
