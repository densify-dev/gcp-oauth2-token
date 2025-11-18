package gcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/densify-dev/gcp-oauth2-token/pkg/k8s"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	tokenFileEnvVar               = "BEARER_TOKEN_FILE" //nolint:gosec
	googleServiceAccountKeyEnvVar = "GOOGLE_SERVICE_ACCOUNT_KEY"
	googleMonitoringReadScope     = "https://www.googleapis.com/auth/monitoring.read"
	tokenFilePerms                = 0600
)

func CreateTokenFile(logger zerolog.Logger) (err error) {
	var tokenFile, googleServiceAccountKey string
	if tokenFile, googleServiceAccountKey, err = validateEnvVars(logger); err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	var tsrc oauth2.TokenSource
	if tsrc, err = tokenSource(ctx, googleServiceAccountKey, googleMonitoringReadScope); err != nil {
		return
	}
	var tok *oauth2.Token
	if tok, err = tsrc.Token(); err != nil {
		return
	}
	if tok.AccessToken == "" {
		err = errors.New("empty access token")
		return
	}
	return os.WriteFile(tokenFile, []byte(tok.AccessToken), tokenFilePerms)
}

func tokenSource(ctx context.Context, googleServiceAccountKey, scope string) (tsrc oauth2.TokenSource, err error) {
	var creds *google.Credentials
	if googleServiceAccountKey != "" {
		var serviceAccountKey []byte
		serviceAccountKey, err = os.ReadFile(googleServiceAccountKey) //nolint:gosec
		if err != nil && len(serviceAccountKey) == 0 {
			err = fmt.Errorf("empty Google service account key: %s", googleServiceAccountKey)
		}
		if err == nil {
			creds, err = google.CredentialsFromJSON(ctx, serviceAccountKey, scope)
		}
	} else {
		creds, err = google.FindDefaultCredentials(ctx, scope)
	}
	if err == nil {
		tsrc = creds.TokenSource
	}
	return
}

func validateEnvVars(logger zerolog.Logger) (tokenFile, googleServiceAccountKey string, err error) {
	tokenFile = os.Getenv(tokenFileEnvVar)
	googleServiceAccountKey = os.Getenv(googleServiceAccountKeyEnvVar)
	if tokenFile == "" {
		err = fmt.Errorf("%s environment variable not set", tokenFileEnvVar)
		return
	}
	_, infoErr := os.Stat(tokenFile)
	switch {
	case infoErr == nil:
		err = fmt.Errorf("%s file already exists", tokenFile)
	case !os.IsNotExist(infoErr):
		err = fmt.Errorf("unable to stat %s: %w", tokenFile, infoErr)
	}
	if err != nil {
		return
	}
	if !k8s.IsNonDefaultServiceAccount(logger) && googleServiceAccountKey == "" {
		err = fmt.Errorf("%s environment variable not set, required for non-Kubernetes environments or if k8s service account is default or its token is not auto-mounted", googleServiceAccountKeyEnvVar)
	}
	return
}
