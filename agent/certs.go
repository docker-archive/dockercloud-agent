package agent

import (
	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/tutumcloud/dockercloud-agent/utils"
)

func CreateCerts(keyFilePath, certFilePath, host string) {
	if !isCertificateExist(keyFilePath, certFilePath) {
		if host == "" {
			os.RemoveAll(AgentPidFile)
			Logger.Fatal("CertCommonName is empty. This may be caused by a failed node registration with Docker Cloud")
		}
		genCetificate(keyFilePath, certFilePath, host)
		Logger.Println("New TLS certificates generated")
	}
}

func isCertificateExist(keyFilePath, certFilePath string) (isExist bool) {
	if utils.FileExist(keyFilePath) && utils.FileExist(certFilePath) {
		return true
	}
	return false
}

func genCetificate(keyFilePath, certFilePath, host string) {
	validFor := 10 * 365 * 24 * time.Hour
	isCA := true
	rsaBits := 2048

	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		SendError(err, "Fatal: Failed to generate private key", nil)
		os.RemoveAll(AgentPidFile)
		Logger.Fatalf("Failed to generate private key: %s", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		SendError(err, "Fatal: Failed to generate serial number", nil)
		os.RemoveAll(AgentPidFile)
		Logger.Fatalf("Failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Docker Cloud Self-Signed Host"},
			CommonName:   host,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		SendError(err, "Fatal: Failed to create certificate", nil)
		os.RemoveAll(AgentPidFile)
		Logger.Fatalf("Failed to create certificate: %s", err)
	}

	certOut, err := os.Create(certFilePath)
	if err != nil {
		SendError(err, "Fatal: Failed to open cert.pem for writing", nil)
		os.RemoveAll(AgentPidFile)
		Logger.Fatalf("Failed to open cert.pem for writing: %s", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(keyFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		SendError(err, "Fatal: Failed to open key.pem for writing", nil)
		os.RemoveAll(AgentPidFile)
		Logger.Fatalf("Failed to open key.pem for writing:", err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
}

func GetCertificate(certFilePath string) (*string, error) {
	content, err := ioutil.ReadFile(certFilePath)
	if err != nil {
		return nil, err
	}
	cert := string(content[:])
	return &cert, nil
}
