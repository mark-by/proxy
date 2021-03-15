package application

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

func crtKeyFilenames(name string) (cert string, key string) {
	return fmt.Sprintf("certs/%s.crt", name), fmt.Sprintf("certs/%s.key", name)
}

func lookUp(name string) (cert string, key string, err error) {
	cert, key = crtKeyFilenames(name)
	_, err = tls.LoadX509KeyPair(cert, key)
	return cert, key, err
}

func certificateTemplate(name string) (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	lastWeek := time.Now().AddDate(0, 0, -7)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"markby"},
			OrganizationalUnit: []string{"markby"},
			CommonName:         name,
		},
		NotBefore:             lastWeek,
		NotAfter:              lastWeek.AddDate(2, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	if ip := net.ParseIP(name); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, name)
	}

	return &template, nil
}

func bigIntHash(n *big.Int) []byte {
	h := sha1.New()
	h.Write(n.Bytes())
	return h.Sum(nil)
}

func createKeyPair(name string) (cert string, key string, err error) {
	template, err := certificateTemplate(name)
	if err != nil {
		return "", "", err
	}

	rootCA, err := tls.LoadX509KeyPair("root.crt", "root.key")
	if err != nil {
		return "", "", err
	}

	if rootCA.Leaf, err = x509.ParseCertificate(rootCA.Certificate[0]); err != nil {
		return "", "", err
	}

	template.AuthorityKeyId = rootCA.Leaf.SubjectKeyId
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template.SubjectKeyId = bigIntHash(privateKey.N)

	certificate, err := x509.CreateCertificate(rand.Reader, template, rootCA.Leaf,
		&privateKey.PublicKey, rootCA.PrivateKey)
	if err != nil {
		return "", "", err
	}

	cert, key = crtKeyFilenames(name)
	certFile, err := os.Create(cert)
	if err != nil {
		return "", "", err
	}
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certificate})
	if err != nil {
		return "", "", err
	}

	keyFile, err := os.OpenFile(key, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", err
	}
	defer keyFile.Close()

	err = pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err != nil {
		return "", "", err
	}

	return cert, key, nil
}

func getCertificateByName(name string) (*tls.Certificate, error) {
	cert, key, err := lookUp(name)
	if err != nil {
		cert, key, err = createKeyPair(name)
		if err != nil {
			return nil, err
		}
	}

	certificate, err := tls.LoadX509KeyPair(cert, key)
	return &certificate, nil
}
