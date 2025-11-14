package gmailfilter

import (
	"context"

	"google.golang.org/api/gmail/v1"
)

// Config is the configuration structure used to instantiate the Google
// provider.
type Config struct {
	gmailService *gmail.Service
}

func (c *Config) LoadAndValidate(ctx context.Context) error {
	gmailService, err := gmail.NewService(ctx)
	if err != nil {
		return err
	}
	c.gmailService = gmailService
	return nil
}
