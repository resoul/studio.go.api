package infrastructure

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type UserTokenManager struct {
	secret []byte
	ttl    time.Duration
}

type userTokenClaims struct {
	UserID uint  `json:"uid"`
	Exp    int64 `json:"exp"`
}

func NewUserTokenManager(secret string, ttl time.Duration) *UserTokenManager {
	return &UserTokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *UserTokenManager) Generate(userID uint) (string, error) {
	claims := userTokenClaims{
		UserID: userID,
		Exp:    time.Now().UTC().Add(m.ttl).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := m.sign(encodedPayload)
	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	return encodedPayload + "." + encodedSignature, nil
}

func (m *UserTokenManager) Parse(token string) (uint, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid token format")
	}

	payloadPart := parts[0]
	signaturePart := parts[1]

	expectedSig := m.sign(payloadPart)
	givenSig, err := base64.RawURLEncoding.DecodeString(signaturePart)
	if err != nil {
		return 0, fmt.Errorf("invalid token signature")
	}

	if subtle.ConstantTimeCompare(expectedSig, givenSig) != 1 {
		return 0, fmt.Errorf("invalid token signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return 0, fmt.Errorf("invalid token payload")
	}

	var claims userTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, fmt.Errorf("invalid token claims")
	}

	if claims.UserID == 0 {
		return 0, fmt.Errorf("invalid token user")
	}

	if time.Now().UTC().Unix() > claims.Exp {
		return 0, fmt.Errorf("token expired")
	}

	return claims.UserID, nil
}

func (m *UserTokenManager) sign(payload string) []byte {
	h := hmac.New(sha256.New, m.secret)
	_, _ = h.Write([]byte(payload))
	return h.Sum(nil)
}
