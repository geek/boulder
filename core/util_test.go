// Copyright 2014 ISRG.  All rights reserved
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package core

import (
	"encoding/json"
	"fmt"
	"github.com/letsencrypt/boulder/Godeps/_workspace/src/github.com/square/go-jose"
	"github.com/letsencrypt/boulder/test"
	"math"
	"math/big"
	"net/url"
	"testing"
)

// challenges.go
func TestNewToken(t *testing.T) {
	token := NewToken()
	fmt.Println(token)
	tokenLength := int(math.Ceil(32 * 8 / 6.0)) // 32 bytes, b64 encoded
	test.AssertIntEquals(t, len(token), tokenLength)
	collider := map[string]bool{}
	// Test for very blatant RNG failures:
	// Try 2^20 birthdays in a 2^72 search space...
	// our naive collision probability here is  2^-32...
	for i := 0; i < 1000000; i++ {
		token = NewToken()[:12] // just sample a portion
		test.Assert(t, !collider[token], "Token collision!")
		collider[token] = true
	}
	return
}

func TestSerialUtils(t *testing.T) {
	serial := SerialToString(big.NewInt(100000000000000000))
	test.AssertEquals(t, serial, "0000000000000000016345785d8a0000")

	serialNum, err := StringToSerial("0000000000000000016345785d8a0000")
	test.AssertNotError(t, err, "Couldn't convert serial number to *big.Int")
	test.AssertBigIntEquals(t, serialNum, big.NewInt(100000000000000000))

	badSerial, err := StringToSerial("doop!!!!000")
	test.AssertEquals(t, fmt.Sprintf("%v", err), "Serial number should be 32 characters long")
	fmt.Println(badSerial)
}

func TestBuildID(t *testing.T) {
	test.AssertEquals(t, "Unspecified", GetBuildID())
}

const JWK_1_JSON = `{
  "kty": "RSA",
  "n": "vuc785P8lBj3fUxyZchF_uZw6WtbxcorqgTyq-qapF5lrO1U82Tp93rpXlmctj6fyFHBVVB5aXnUHJ7LZeVPod7Wnfl8p5OyhlHQHC8BnzdzCqCMKmWZNX5DtETDId0qzU7dPzh0LP0idt5buU7L9QNaabChw3nnaL47iu_1Di5Wp264p2TwACeedv2hfRDjDlJmaQXuS8Rtv9GnRWyC9JBu7XmGvGDziumnJH7Hyzh3VNu-kSPQD3vuAFgMZS6uUzOztCkT0fpOalZI6hqxtWLvXUMj-crXrn-Maavz8qRhpAyp5kcYk3jiHGgQIi7QSK2JIdRJ8APyX9HlmTN5AQ",
  "e": "AAEAAQ"
}`
const JWK_1_DIGEST = `ul04Iq07ulKnnrebv2hv3yxCGgVvoHs8hjq2tVKx3mc=`
const JWK_2_JSON = `{
  "kty":"RSA",
  "n":"yTsLkI8n4lg9UuSKNRC0UPHsVjNdCYk8rGXIqeb_rRYaEev3D9-kxXY8HrYfGkVt5CiIVJ-n2t50BKT8oBEMuilmypSQqJw0pCgtUm-e6Z0Eg3Ly6DMXFlycyikegiZ0b-rVX7i5OCEZRDkENAYwFNX4G7NNCwEZcH7HUMUmty9dchAqDS9YWzPh_dde1A9oy9JMH07nRGDcOzIh1rCPwc71nwfPPYeeS4tTvkjanjeigOYBFkBLQuv7iBB4LPozsGF1XdoKiIIi-8ye44McdhOTPDcQp3xKxj89aO02pQhBECv61rmbPinvjMG9DYxJmZvjsKF4bN2oy0DxdC1jDw",
  "e":"AAEAAQ"
}`

func TestKeyDigest(t *testing.T) {
	// Test with JWK (value, reference, and direct)
	var jwk jose.JsonWebKey
	json.Unmarshal([]byte(JWK_1_JSON), &jwk)
	digest, err := KeyDigest(jwk)
	test.Assert(t, err == nil && digest == JWK_1_DIGEST, "Failed to digest JWK by value")
	digest, err = KeyDigest(&jwk)
	test.Assert(t, err == nil && digest == JWK_1_DIGEST, "Failed to digest JWK by reference")
	digest, err = KeyDigest(jwk.Key)
	test.Assert(t, err == nil && digest == JWK_1_DIGEST, "Failed to digest bare key")

	// Test with unknown key type
	digest, err = KeyDigest(struct{}{})
	test.Assert(t, err != nil, "Should have rejected unknown key type")
}

func TestKeyDigestEquals(t *testing.T) {
	var jwk1, jwk2 jose.JsonWebKey
	json.Unmarshal([]byte(JWK_1_JSON), &jwk1)
	json.Unmarshal([]byte(JWK_2_JSON), &jwk2)

	test.Assert(t, KeyDigestEquals(jwk1, jwk1), "Key digests for same key should match")
	test.Assert(t, !KeyDigestEquals(jwk1, jwk2), "Key digests for different keys should not match")
	test.Assert(t, !KeyDigestEquals(jwk1, struct{}{}), "Unknown key types should not match anything")
	test.Assert(t, !KeyDigestEquals(struct{}{}, struct{}{}), "Unknown key types should not match anything")
}

func TestFingerprintEquals(t *testing.T) {
	buf := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	fp := []byte{0x8a, 0x85, 0x1f, 0xf8, 0x2e, 0xe7, 0x04, 0x8a,
		0xd0, 0x9e, 0xc3, 0x84, 0x7f, 0x1d, 0xdf, 0x44,
		0x94, 0x41, 0x04, 0xd2, 0xcb, 0xd1, 0x7e, 0xf4,
		0xe3, 0xdb, 0x22, 0xc6, 0x78, 0x5a, 0x0d, 0x45}
	test.Assert(t, FingerprintEquals(buf, fp), "Fingerprint did not match")
}

func TestAcmeURL(t *testing.T) {
	s := "http://example.invalid"
	u, _ := url.Parse(s)
	a := AcmeURL(*u)
	test.AssertEquals(t, s, a.String())
}
