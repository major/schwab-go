package auth

import (
	"crypto/x509"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	t.Parallel()

	certificate, err := generateSelfSignedCert()
	require.NoError(t, err)
	require.NotEmpty(t, certificate.Certificate)

	parsedCert, err := x509.ParseCertificate(certificate.Certificate[0])
	require.NoError(t, err)

	t.Run("valid only for loopback ip", func(t *testing.T) {
		t.Parallel()

		require.Len(t, parsedCert.IPAddresses, 1)
		assert.True(t, parsedCert.IPAddresses[0].Equal(net.ParseIP("127.0.0.1")))
	})

	t.Run("not valid for other ip", func(t *testing.T) {
		t.Parallel()

		err := parsedCert.VerifyHostname("192.168.1.1")
		require.Error(t, err)
	})

	t.Run("self signed", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, parsedCert.Subject.String(), parsedCert.Issuer.String())
	})

	t.Run("server auth extended key usage", func(t *testing.T) {
		t.Parallel()

		assert.Contains(t, parsedCert.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	})
}
