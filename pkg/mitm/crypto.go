package mitm

import (
	"crypto/tls"
	"fmt"
	"time"
)

const (
	ONE_DAY   = 24 * time.Hour
	TWO_WEEKS = ONE_DAY * 14
	ONE_MONTH = 1
	ONE_YEAR  = 1
	TEN_YEAR  = 10 * ONE_YEAR
)

// CryptoConfig configures the cryptography settings for an MITMer
type CryptoConfig struct {
	// PKFile: the PEM-encoded file to use as the primary key for this server
	PKFile string

	// CertFile: the PEM-encoded X509 certificate to use for this server (must match PKFile)
	CertFile string

	// Organization: Name of the organization to use on the generated CA cert for this  (defaults to "gomitm")
	Organization string

	// CommonName: CommonName to use on the generated CA cert for this proxy (defaults to "Lantern")
	CommonName string

	// ServerTLSConfig: optional configuration for TLS server when MITMing (if nil, a sensible default is used)
	ServerTLSConfig *tls.Config
}

func (wrapper *HandlerWrapper) initCrypto() (err error) {
	if wrapper.cryptoConf.Organization == "" {
		wrapper.cryptoConf.Organization = "scfproxy"
	}
	if wrapper.cryptoConf.CommonName == "" {
		wrapper.cryptoConf.CommonName = "Lantern"
	}
	if wrapper.pk, err = LoadPKFromFile(wrapper.cryptoConf.PKFile); err != nil {
		wrapper.pk, err = GeneratePK(2048)
		if err != nil {
			return fmt.Errorf("Unable to generate private key: %s", err)
		}
		wrapper.pk.WriteToFile(wrapper.cryptoConf.PKFile)
	}
	wrapper.pkPem = wrapper.pk.PEMEncoded()
	wrapper.issuingCert, err = LoadCertificateFromFile(wrapper.cryptoConf.CertFile)
	if err != nil || wrapper.issuingCert.ExpiresBefore(time.Now().AddDate(0, ONE_MONTH, 0)) {
		wrapper.issuingCert, err = wrapper.pk.TLSCertificateFor(
			wrapper.cryptoConf.Organization,
			wrapper.cryptoConf.CommonName,
			time.Now().AddDate(TEN_YEAR, 0, 0),
			true,
			nil)
		if err != nil {
			return fmt.Errorf("Unable to generate self-signed issuing certificate: %s", err)
		}
		wrapper.issuingCert.WriteToFile(wrapper.cryptoConf.CertFile)
	}
	wrapper.issuingCertPem = wrapper.issuingCert.PEMEncoded()
	return
}

func (wrapper *HandlerWrapper) mitmCertForName(name string) (cert *tls.Certificate, err error) {
	// Try to read an existing cert
	kpCandidateIf, found := wrapper.dynamicCerts.Get(name)
	if found {
		return kpCandidateIf.(*tls.Certificate), nil
	}

	// Existing cert not found, lock for writing and recheck
	wrapper.certMutex.Lock()
	defer wrapper.certMutex.Unlock()
	kpCandidateIf, found = wrapper.dynamicCerts.Get(name)
	if found {
		return kpCandidateIf.(*tls.Certificate), nil
	}

	// Still not found, create certificate
	certTTL := TWO_WEEKS
	generatedCert, err := wrapper.pk.TLSCertificateFor(
		wrapper.cryptoConf.Organization,
		name,
		time.Now().Add(certTTL),
		false,
		wrapper.issuingCert)
	if err != nil {
		return nil, fmt.Errorf("Unable to issue certificate: %s", err)
	}
	keyPair, err := tls.X509KeyPair(generatedCert.PEMEncoded(), wrapper.pkPem)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse keypair for tls: %s", err)
	}

	// Add to cache, set to expire 1 day before the cert expires
	cacheTTL := certTTL - ONE_DAY
	wrapper.dynamicCerts.Set(name, &keyPair, cacheTTL)
	return &keyPair, nil
}
