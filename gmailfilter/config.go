package gmailfilter

import (
	"context"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Config is the configuration structure used to instantiate the Google
// provider.
type Config struct {
	client           *http.Client
	gmailService     *gmail.Service
	context          context.Context
	terraformVersion string
	userAgent        string

	tokenSource oauth2.TokenSource
}

func (c *Config) LoadAndValidate(ctx context.Context) error {

	gmailService, err := gmail.NewService(ctx)
	if err != nil {
		return nil
	}
	c.gmailService = gmailService
	return nil
}

func (c *Config) getTokenSource(clientScopes []string) (oauth2.TokenSource, error) {
	// Use Application Default Credentials (ADC)
	log.Printf("[INFO] Authenticating using Application Default Credentials (ADC)...")
	log.Printf("[INFO]   -- Scopes: %s", clientScopes)
	return googleoauth.DefaultTokenSource(context.Background(), clientScopes...)
}
