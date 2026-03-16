package main
import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)
func main() {
	err := bcrypt.CompareHashAndPassword([]byte("$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS"), []byte("demo123"))
	fmt.Printf("Matches: %v\n", err == nil)
}
