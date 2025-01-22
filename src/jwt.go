package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

type CustomClaims struct {
	UID        string `json:"uid"`
	Permission int    `json:"permission"`
	jwt.StandardClaims
}

func ParseKeys(rawSK []byte, rawPK []byte) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	var err error
	privateKeyBlock, _ := pem.Decode(rawSK)
	if privateKeyBlock == nil {
		return nil, nil, errors.New("private key cannot decode")
	}
	var secretKey *rsa.PrivateKey
	if privateKeyBlock.Type == "RSA PRIVATE KEY" {
		secretKey, err = x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			return nil, nil, errors.New("failed to parse private key")
		}
	} else if privateKeyBlock.Type == "PRIVATE KEY" {
		key, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			return nil, nil, errors.New("failed to parse private key")
		}
		var ok bool
		secretKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, nil, errors.New("private key is not of type *rsa.PrivateKey")
		}
	} else {
		return nil, nil, errors.New("private key type is invalid")
	}

	publicKeyBlock, _ := pem.Decode(rawPK)
	if publicKeyBlock == nil {
		return nil, nil, errors.New("public key cannot decode")
	}
	if publicKeyBlock.Type != "PUBLIC KEY" {
		return nil, nil, errors.New("public key type is invalid")
	}

	pubKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, nil, errors.New("failed to parse public key")
	}
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, nil, errors.New("public key is not of type *rsa.PublicKey")
	}
	return secretKey, rsaPubKey, nil
}

func GenerateJWT(uid string, permission int, secretKey *rsa.PrivateKey) (string, error) {
	claims := CustomClaims{
		UID:        uid,
		Permission: permission,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(secretKey)
}

func VerifyToken(tokenString string, publicKey *rsa.PublicKey) (*jwt.Token, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// check signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			err := errors.New("Unexpected signing method")
			return nil, err
		}
		return publicKey, nil
	})
	if err != nil {
		err = errors.Wrap(err, "Token is invalid")
		return nil, err
	}
	if !parsedToken.Valid {
		return nil, errors.New("Token is invalid")
	}

	return parsedToken, nil
}
