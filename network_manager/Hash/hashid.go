package hashid

import (
	"crypto/sha256"
	"encoding/hex"
	"log"

	"github.com/denisbrodbeck/machineid"
)

// GetHashedUUID returns the hashed UUID of the machine.
func GetHashedUUID() string {
	// Get the machine UUID
	id, err := machineid.ID()
	if err != nil {
		log.Fatal(err)
	}

	// Hash the UUID using SHA-256
	hash := sha256.New()
	hash.Write([]byte(id))

	// Return the hex representation of the hash
	return hex.EncodeToString(hash.Sum(nil))
}
