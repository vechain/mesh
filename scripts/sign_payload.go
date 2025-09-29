package main

import (
	"fmt"
	"os"

	meshcrypto "github.com/vechain/mesh/common/crypto"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <private_key_hex> <payload_hex>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s 99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36 c7c260e16e3c32a6176759a3556ff5618d7d6e7e2c9c9602d40461fcaa34cbec\n", os.Args[0])
		os.Exit(1)
	}

	privateKeyHex := os.Args[1]
	payloadHex := os.Args[2]

	// Use the common signing functionality
	signature, address, err := meshcrypto.NewSigningHandler(privateKeyHex).SignPayloadWithAddress(payloadHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Output the signature
	fmt.Println(signature)

	// Show derived address
	fmt.Fprintf(os.Stderr, "Derived address: %s\n", address)
}
