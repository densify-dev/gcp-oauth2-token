package k8s

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
)

const (
	ServiceAccountTokenFileEnvVar = "SERVICE_ACCOUNT_TOKEN_FILE"
	// DefaultServiceAccountTokenFile is the default path where the service account token
	// is mounted in a Kubernetes pod.
	DefaultServiceAccountTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token" //nolint:gosec
	Default                        = "default"
)

// TokenClaims embeds the standard JWT claims and adds the custom Kubernetes-specific claims.
type TokenClaims struct {
	jwt.RegisteredClaims
	TokenInfo TokenInfo `json:"kubernetes.io"`
}

// TokenInfo holds all the claims nested under the "kubernetes.io" key.
type TokenInfo struct {
	Namespace      string           `json:"namespace"`
	Node           ObjectInfo       `json:"node"`
	Pod            ObjectInfo       `json:"pod"`
	ServiceAccount ObjectInfo       `json:"serviceaccount"`
	WarnAfter      *jwt.NumericDate `json:"warnafter,omitempty"`
}

// ObjectInfo contains details about a k8s object.
type ObjectInfo struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

func (ti TokenInfo) IsDefaultServiceAccount() bool {
	return ti.ServiceAccount.Name == Default
}

func IsNonDefaultServiceAccount(logger zerolog.Logger) bool {
	// Read the token from the file.
	tokenBytes, err := os.ReadFile(getServiceAccountTokenFileName())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Info().Err(err).Msg("Not running in a Kubernetes pod or the token is not auto-mounted")
		} else {
			logger.Error().Err(err).Msg("Error reading token file")
		}
		return false
	}
	claims := &TokenClaims{}
	// Parse the token without validating the signature.
	// We pass a nil Keyfunc because we don't have the key to validate the signature,
	// and for this use case, we trust the token file mounted by Kubernetes.
	_, err = jwt.ParseWithClaims(string(tokenBytes), claims, nil)
	// The library will return an error when no key is provided, even if parsing succeeds.
	// We check if the error is the specific one for missing validation, which we can ignore.
	if err != nil && !errors.Is(err, jwt.ErrTokenUnverifiable) {
		logger.Error().Err(err).Msg("Error parsing token")
		return false
	}
	logger.Debug().Msgf("Token claims: %+v", claims)
	return !claims.TokenInfo.IsDefaultServiceAccount()
}

func getServiceAccountTokenFileName() string {
	fileName := os.Getenv(ServiceAccountTokenFileEnvVar)
	if fileName == "" {
		fileName = DefaultServiceAccountTokenFile
	}
	return fileName
}
