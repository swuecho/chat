package auth

import (
	"fmt"
	"strings"
	"testing"
)

func TestGeneratePasswordHash(t *testing.T) {
	password := "mypassword"

	hash, err := GeneratePasswordHash(password)
	if err != nil {
		t.Fatalf("error generating password hash: %v", err)
	}
	fmt.Println(hash)
	// Check that the hash has the correct format
	fields := strings.Split(hash, "$")
	if len(fields) != 4 || fields[0] != "pbkdf2_sha256" || fields[1] != "260000" {
		t.Errorf("unexpected hash format: %s", hash)
	}

	// Check that we can successfully validate the password using the hash
	valid := ValidatePassword(password, hash)
	if !valid {
		t.Error("generated hash does not validate password")
	}
}

func TestGeneratePasswordHash2(t *testing.T) {
	password := "@WuHao5"

	hash, err := GeneratePasswordHash(password)
	if err != nil {
		t.Fatalf("error generating password hash: %v", err)
	}
	fmt.Println(hash)
	// Check that the hash has the correct format
	fields := strings.Split(hash, "$")
	if len(fields) != 4 || fields[0] != "pbkdf2_sha256" || fields[1] != "260000" {
		t.Errorf("unexpected hash format: %s", hash)
	}

	// Check that we can successfully validate the password using the hash
	valid := ValidatePassword(password, hash)
	if !valid {
		t.Error("generated hash does not validate password")
	}
}

func TestPass(t *testing.T) {
	hash := "pbkdf2_sha256$260000$TSefBGfPi5fY+4whotY5sQ==$/1CeWE2PG6aYdW2DSxYyVol+HEZBmAfDj7zMgEMlxgg="
	password := "using555"
	// Check that we can successfully validate the password using the hash
	valid := ValidatePassword(password, hash)
	if !valid {
		t.Error("generated hash does not validate password")
	}

}
