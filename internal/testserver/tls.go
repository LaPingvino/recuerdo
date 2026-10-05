package testserver

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Certificate makes a self-signed certificate for the test server in dir
// the first time (testserver-cert.pem, testserver-key.pem; for this
// computer's names and addresses) and reuses it afterwards. Clients
// check its fingerprint (SHA-256), which the server prints.
func Certificate(dir string) (certFile, keyFile, fingerprint string, err error) {
	certFile, keyFile = filepath.Join(dir, "testserver-cert.pem"), filepath.Join(dir, "testserver-key.pem")
	if _, err := os.Stat(certFile); err != nil {
		if err := makeCertificate(certFile, keyFile); err != nil {
			return "", "", "", err
		}
	}
	data, err := os.ReadFile(certFile)
	if err != nil {
		return "", "", "", err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return "", "", "", fmt.Errorf("%s is not a certificate", certFile)
	}
	return certFile, keyFile, Fingerprint(block.Bytes), nil
}

// Fingerprint is a certificate's SHA-256 as AA:BB:...
func Fingerprint(der []byte) string {
	sum := sha256.Sum256(der)
	parts := make([]string, len(sum))
	for i, b := range sum {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(parts, ":")
}

func makeCertificate(certFile, keyFile string) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 127))
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "Recuerdo test server", Organization: []string{"Recuerdo"}},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}
	if host, err := os.Hostname(); err == nil && host != "" {
		tmpl.DNSNames = append(tmpl.DNSNames, host, host+".local")
	}
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok && !ipn.IP.IsLoopback() && !ipn.IP.IsLinkLocalUnicast() {
				tmpl.IPAddresses = append(tmpl.IPAddresses, ipn.IP)
			}
		}
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(certFile), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o600); err != nil {
		return err
	}
	return os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644)
}
