package methods

import (
	"crypto/x509"
	"strings"

	"google.golang.org/grpc/credentials"
)

func getTLSConfig(endpoint string) credentials.TransportCredentials {
	if strings.HasPrefix(endpoint, "https://") || (!strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "localhost")) {
		pool, _ := x509.SystemCertPool()
		return credentials.NewClientTLSFromCert(pool, "")
	}
	return nil
}
