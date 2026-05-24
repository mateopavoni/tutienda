// Package authx holds the auth primitives shared by the services: a minimal HS256 JWT (no external
// dependency, kept deliberately small and auditable, in the spirit of the dependency-free config) and
// bcrypt password hashing.
package authx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ErrInvalidToken is returned for any malformed, tampered or expired token.
var ErrInvalidToken = errors.New("invalid token")

// Platform roles. RoleMerchant is the default (owns and manages their own stores); RoleAdmin is the
// SaaS super-admin who manages every tenant cross-tenant. The role is a property of the merchant and
// rides only the merchant-scoped token (the gateway gates /api/platform on it).
const (
	RoleMerchant = "merchant"
	RoleAdmin    = "admin"
)

// Claims is the JWT payload. MerchantID owns the session; StoreID is the active tenant the token is
// scoped to (empty on a merchant-scoped token); Plan is that store's membership tier, signed in so the
// gateway can gate paid features without a DB lookup; Role is the platform role on the merchant token.
type Claims struct {
	MerchantID string `json:"sub"`
	StoreID    string `json:"store,omitempty"`
	Plan       string `json:"plan,omitempty"`
	Role       string `json:"role,omitempty"`
	ExpiresAt  int64  `json:"exp"`
}

// header is the fixed JOSE header for HS256; we only ever issue this one algorithm.
var header = base64url([]byte(`{"alg":"HS256","typ":"JWT"}`))

// Issuer signs and verifies tokens with a single shared secret (HS256).
type Issuer struct {
	secret []byte
	ttl    time.Duration
}

// NewIssuer builds an issuer; ttl is how long a freshly issued token stays valid.
func NewIssuer(secret string, ttl time.Duration) *Issuer {
	return &Issuer{secret: []byte(secret), ttl: ttl}
}

// IssueMerchant signs a merchant-scoped token: no store, carrying the platform role so the gateway can
// gate the cross-tenant /api/platform routes. Used by signup and login.
func (i *Issuer) IssueMerchant(merchantID, role string) (string, error) {
	return i.issue(Claims{MerchantID: merchantID, Role: role})
}

// IssueStore signs a store-scoped token: the active tenant and its membership plan. Used after a
// merchant selects a store; ownership is verified before this is called.
func (i *Issuer) IssueStore(merchantID, storeID, plan string) (string, error) {
	return i.issue(Claims{MerchantID: merchantID, StoreID: storeID, Plan: plan})
}

// issue stamps the expiry and signs the claims.
func (i *Issuer) issue(claims Claims) (string, error) {
	claims.ExpiresAt = time.Now().Add(i.ttl).Unix()
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signing := header + "." + base64url(payloadJSON)
	return signing + "." + base64url(i.sign(signing)), nil
}

// Verify checks the signature and expiry and returns the decoded claims.
func (i *Issuer) Verify(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	signing := parts[0] + "." + parts[1]
	want := i.sign(signing)
	got, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(want, got) {
		return nil, ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().Unix() >= claims.ExpiresAt {
		return nil, ErrInvalidToken
	}
	return &claims, nil
}

func (i *Issuer) sign(signing string) []byte {
	mac := hmac.New(sha256.New, i.secret)
	mac.Write([]byte(signing))
	return mac.Sum(nil)
}

func base64url(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
