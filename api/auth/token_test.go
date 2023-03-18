package auth

import (
	"fmt"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	user_id := int32(0)

	token, err := GenerateToken(user_id)
	if err != nil {
		t.Fatalf("error generating password hash: %v", err)
	}
	// Check that the hash has the correct format
	// Check that we can successfully validate the password using the hash
	fmt.Println(token)
	user_id_after_valid, err := ValidateToken(token)
	if err != nil {
		t.Error("generated token does not validate ")
	}
	if user_id != user_id_after_valid {
		t.Error("generated token does not validate ")
	}
}
