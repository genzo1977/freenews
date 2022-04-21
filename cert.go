package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

func setupCerts() {

	var err error
	var caPrivKey *rsa.PrivateKey
	var caPEMBuffer *bytes.Buffer

	if _, err := os.Stat("cert/ca.pem"); err == nil {

		// pem decode
		caPEMBytes, err := os.ReadFile("cert/ca.pem")
		if err != nil {
			log.Fatal(err)
		}
		caString = string(caPEMBytes)
		caPEMBuffer = bytes.NewBuffer(caPEMBytes)
		caPEM, _ := pem.Decode(caPEMBytes)
		ca, err = x509.ParseCertificate(caPEM.Bytes)

		caPrivKeyPEMBytes, err := os.ReadFile("cert/key.pem")
		if err != nil {
			log.Fatal(err)
		}
		caPrivKeyPEM, _ := pem.Decode(caPrivKeyPEMBytes)
		caPrivKey, err = x509.ParsePKCS1PrivateKey(caPrivKeyPEM.Bytes)

	} else if errors.Is(err, os.ErrNotExist) {

		ca = &x509.Certificate{
			SerialNumber: big.NewInt(2019),
			Subject: pkix.Name{
				Organization:  []string{"Freenews Org"},
				Country:       []string{"US"},
				Province:      []string{""},
				Locality:      []string{"San Francisco"},
				StreetAddress: []string{"Golden Gate Bridge"},
				PostalCode:    []string{"94016"},
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().AddDate(10, 0, 0),
			IsCA:                  true,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
		}

		// create our private and public key
		caPrivKey, err = rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			log.Fatal(err)
		}

		// create the CA
		caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
		if err != nil {
			log.Fatal(err)
		}

		// pem encode
		caPEMBuffer = new(bytes.Buffer)
		pem.Encode(caPEMBuffer, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caBytes,
		})
		caString = caPEMBuffer.String()
		if err = os.WriteFile("cert/ca.pem", []byte(caString), 0600); err != nil {
			log.Fatal(err)
		}

		caPrivKeyPEMBuffer := new(bytes.Buffer)
		pem.Encode(caPrivKeyPEMBuffer, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
		})
		caPrivKeyString := caPrivKeyPEMBuffer.String()
		if err = os.WriteFile("cert/key.pem", []byte(caPrivKeyString), 0600); err != nil {
			log.Fatal(err)
		}
	}

	var dnsNames []string
	for _, host := range proxyHosts {
		dnsNames = append(dnsNames, fmt.Sprintf("*.%s", host))
		dnsNames = append(dnsNames, host)
	}

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization:  []string{"Freenews"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		//IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     dnsNames,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		log.Fatal(err)
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	tlsHttpServerConfig = &tls.Config{
		MinVersion:               tls.VersionTLS10,
		NextProtos:               []string{"http/1.1"},
		Certificates:             []tls.Certificate{serverCert},
		PreferServerCipherSuites: true,
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(caPEMBuffer.Bytes())
}

func setupDoTCerts(){
	if _, err := os.Stat("cert/dot_cert.pem"); err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat("cert/dot_key.pem"); err != nil {
		log.Fatal(err)
	}

	// pem decode
	certPEMBytes, err := os.ReadFile("cert/dot_cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	certPrivKeyPEMBytes, err := os.ReadFile("cert/dot_key.pem")
	if err != nil {
		log.Fatal(err)
	}

	serverCert, err := tls.X509KeyPair(certPEMBytes, certPrivKeyPEMBytes)
	if err != nil {
		log.Fatal(err)
	}

	tlsDoTServerConfig = &tls.Config{
		Certificates:             []tls.Certificate{serverCert},
		PreferServerCipherSuites: true,
	}
	
}