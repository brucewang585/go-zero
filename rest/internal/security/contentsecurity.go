package security

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/brucewang585/go-zero/core/codec"
	"github.com/brucewang585/go-zero/core/iox"
	"github.com/brucewang585/go-zero/core/logx"
	"github.com/brucewang585/go-zero/rest/httpx"
)

const (
	requestUriHeader = "X-Request-Uri"
	signatureField   = "signature"
	timeField        = "time"
)

var (
	ErrInvalidContentType = errors.New("invalid content type")
	ErrInvalidHeader      = errors.New("invalid X-Content-Security header")
	ErrInvalidKey         = errors.New("invalid key")
	ErrInvalidPublicKey   = errors.New("invalid public key")
	ErrInvalidSecret      = errors.New("invalid secret")
)

type ContentSecurityHeader struct {
	Key         []byte
	Timestamp   string
	ContentType int
	Signature   string
}

func (h *ContentSecurityHeader) Encrypted() bool {
	return h.ContentType == httpx.CryptionType
}

func ParseContentSecurity(decrypters map[string]codec.RsaDecrypter, r *http.Request) (
	*ContentSecurityHeader, error) {
	contentSecurity := r.Header.Get(httpx.ContentSecurity)
	attrs := httpx.ParseHeader(contentSecurity)
	fingerprint := attrs[httpx.KeyField]
	secret := attrs[httpx.SecretField]
	signature := attrs[signatureField]

	if len(fingerprint) == 0 || len(secret) == 0 || len(signature) == 0 {
		return nil, ErrInvalidHeader
	}

	decrypter, ok := decrypters[fingerprint]
	if !ok {
		return nil, ErrInvalidPublicKey
	}

	decryptedSecret, err := decrypter.DecryptBase64(secret)
	if err != nil {
		return nil, ErrInvalidSecret
	}

	attrs = httpx.ParseHeader(string(decryptedSecret))
	base64Key := attrs[httpx.KeyField]
	timestamp := attrs[timeField]
	contentType := attrs[httpx.TypeField]

	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, ErrInvalidKey
	}

	cType, err := strconv.Atoi(contentType)
	if err != nil {
		return nil, ErrInvalidContentType
	}

	return &ContentSecurityHeader{
		Key:         key,
		Timestamp:   timestamp,
		ContentType: cType,
		Signature:   signature,
	}, nil
}

func VerifySignature(r *http.Request, securityHeader *ContentSecurityHeader, tolerance time.Duration) int {
	seconds, err := strconv.ParseInt(securityHeader.Timestamp, 10, 64)
	if err != nil {
		return httpx.CodeSignatureInvalidHeader
	}

	now := time.Now().Unix()
	toleranceSeconds := int64(tolerance.Seconds())
	if seconds+toleranceSeconds < now || now+toleranceSeconds < seconds {
		return httpx.CodeSignatureWrongTime
	}

	reqPath, reqQuery := getPathQuery(r)
	signContent := strings.Join([]string{
		securityHeader.Timestamp,
		r.Method,
		reqPath,
		reqQuery,
		computeBodySignature(r),
	}, "\n")
	actualSignature := codec.HmacBase64(securityHeader.Key, signContent)

	passed := securityHeader.Signature == actualSignature
	if !passed {
		logx.Infof("signature different, expect: %s, actual: %s",
			securityHeader.Signature, actualSignature)
	}

	if passed {
		return httpx.CodeSignaturePass
	} else {
		return httpx.CodeSignatureInvalidToken
	}
}

func computeBodySignature(r *http.Request) string {
	var dup io.ReadCloser
	r.Body, dup = iox.DupReadCloser(r.Body)
	sha := sha256.New()
	io.Copy(sha, r.Body)
	r.Body = dup
	return fmt.Sprintf("%x", sha.Sum(nil))
}

func getPathQuery(r *http.Request) (string, string) {
	requestUri := r.Header.Get(requestUriHeader)
	if len(requestUri) == 0 {
		return r.URL.Path, r.URL.RawQuery
	}

	uri, err := url.Parse(requestUri)
	if err != nil {
		return r.URL.Path, r.URL.RawQuery
	}

	return uri.Path, uri.RawQuery
}
