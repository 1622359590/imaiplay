package security

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("HashPassword() stored plaintext")
	}
	if !CheckPassword("correct horse battery staple", hash) {
		t.Fatal("CheckPassword() rejected valid password")
	}
	if CheckPassword("wrong", hash) {
		t.Fatal("CheckPassword() accepted invalid password")
	}
}
