package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "Password123"
	storedHash := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"
	
	fmt.Printf("Testing password: '%s'\n", password)
	fmt.Printf("Against hash: '%s'\n", storedHash)
	
	// Test the stored hash
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		fmt.Printf("❌ VERIFICATION FAILED: %v\n", err)
		fmt.Println("\nThe stored hash does NOT match 'Password123'")
	} else {
		fmt.Printf("✅ VERIFICATION SUCCESS\n")
		fmt.Println("The stored hash DOES match 'Password123'")
	}
	
	// Generate a correct hash for password123
	fmt.Println("\nGenerating correct hash for 'Password123':")
	correctHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		return
	}
	
	fmt.Printf("Correct hash: %s\n", string(correctHash))
	
	// Verify the new hash works
	err = bcrypt.CompareHashAndPassword(correctHash, []byte(password))
	if err != nil {
		fmt.Printf("❌ New hash verification failed: %v\n", err)
	} else {
		fmt.Printf("✅ New hash verification succeeded\n")
	}
}
