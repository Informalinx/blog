package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
)

func main() {
	args := os.Args

	if len(args) <= 1 {
		return
	}

	command := args[1]

	if err := ExecCommand(command, args[1:]); err != nil {
		log.Fatal(err)
	}
}

func ExecCommand(name string, args []string) error {
	switch name {
	case "secret:generate":
		secret, err := GenerateSecret()
		if err != nil {
			return err
		}

		fmt.Println("Generated secret :", secret)
	default:
		return fmt.Errorf("unknown command %q", name)
	}

	return nil
}

func GenerateSecret() (string, error) {
	secret := [64]byte{}
	if _, err := rand.Reader.Read(secret[:]); err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(secret[:]), nil
}
