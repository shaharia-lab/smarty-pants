// Package llm contains the language model provider interface and its implementations.
package llm

// Fake is a fake language model provider
type Fake struct {
}

// HealthCheck checks the health of the language model provider
func (f *Fake) HealthCheck() error {
	return nil
}

// GetResponse returns a response from the language model provider
func (f *Fake) GetResponse(_ Prompt) (string, error) {
	return "", nil
}
