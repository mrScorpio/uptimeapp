package auth

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func TestAccessAndRefresh(t *testing.T) {
	tokens := Tokens{Secret: []byte("a-secret-of-at-least-thirty-two-characters"), AccessTTL: time.Minute}
	raw, err := tokens.Access(42)
	if err != nil {
		t.Fatal(err)
	}
	id, err := tokens.Verify(raw)
	if err != nil || id != 42 {
		t.Fatalf("verify: id=%d err=%v", id, err)
	}
	if _, err := tokens.Verify(raw + "broken"); err == nil {
		t.Fatal("accepted invalid JWT")
	}
	a, hashA, err := Refresh()
	if err != nil {
		t.Fatal(err)
	}
	b, hashB, err := Refresh()
	if err != nil {
		t.Fatal(err)
	}
	if a == b || hashA == hashB || Hash(a) != hashA {
		t.Fatal("refresh tokens are not unique or hashes mismatch")
	}
}

func TestExpiredAccess(t *testing.T) {
	tokens := Tokens{Secret: []byte("a-secret-of-at-least-thirty-two-characters"), AccessTTL: -time.Minute}
	raw, err := tokens.Access(1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tokens.Verify(raw); err == nil {
		t.Fatal("accepted expired JWT")
	}
}

func TestPasswordHash(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte("password123")); err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte("wrong")); err == nil {
		t.Fatal("accepted wrong password")
	}
}
