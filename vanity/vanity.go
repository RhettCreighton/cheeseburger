package vanity

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base32"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"

	"golang.org/x/crypto/sha3"
)

func RunVanity() {
	prefix := flag.String("prefix", "", "vanity prefix for onion address (in lowercase)")
	saveMode := flag.Bool("save", false, "save the generated key information to a file")
	workers := flag.Int("workers", runtime.NumCPU(), "number of parallel workers")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var totalAttempts uint64
	prefixLower := strings.ToLower(*prefix)
	version := byte(0x03)

	type result struct {
		onionAddr string
		finalPub  ed25519.PublicKey
		expanded  [64]byte
		attempts  uint64
	}

	resultChan := make(chan result)

	worker := func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				pub, priv, err := ed25519.GenerateKey(rand.Reader)
				if err != nil {
					log.Fatalf("Failed to generate key pair: %v", err)
				}
				atomic.AddUint64(&totalAttempts, 1)
				expanded := sha512.Sum512(priv[:32])
				expanded[0] &= 248
				expanded[31] &= 127
				expanded[31] |= 64

				checksumData := append(append([]byte(".onion checksum"), pub...), version)
				torHash := sha3.Sum256(checksumData)

				onionBytes := append(append([]byte{}, pub...), torHash[0:2]...)
				onionBytes = append(onionBytes, version)
				encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(onionBytes)
				onionAddr := strings.ToLower(encoded)

				if prefixLower == "" || strings.HasPrefix(onionAddr, prefixLower) {
					resultChan <- result{
						onionAddr: onionAddr,
						finalPub:  pub,
						expanded:  expanded,
						attempts:  atomic.LoadUint64(&totalAttempts),
					}
					return
				}

				if atomic.LoadUint64(&totalAttempts)%1000000 == 0 {
					log.Printf("Total Attempts: %d", atomic.LoadUint64(&totalAttempts))
				}
			}
		}
	}

	for i := 0; i < *workers; i++ {
		go worker()
	}

	res := <-resultChan
	cancel()

	if *saveMode {
		saveDir := filepath.Join("data", "vanity", "default")
		if err := os.MkdirAll(saveDir, 0700); err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}

		secretFilePath := filepath.Join(saveDir, "hs_ed25519_secret_key")
		secretData := append([]byte("== ed25519v1-secret: type0 ==\x00\x00\x00"), res.expanded[:]...)
		if err := os.WriteFile(secretFilePath, secretData, 0600); err != nil {
			log.Fatalf("Failed to write secret key: %v", err)
		}
		log.Printf("Secret key saved to: %s", secretFilePath)

		pubFilePath := filepath.Join(saveDir, "hs_ed25519_public_key")
		pubData := append([]byte("== ed25519v1-public: type0 ==\x00\x00\x00"), res.finalPub...)
		if err := os.WriteFile(pubFilePath, pubData, 0600); err != nil {
			log.Fatalf("Failed to write public key: %v", err)
		}
		log.Printf("Public key saved to: %s", pubFilePath)

		hostnamePath := filepath.Join(saveDir, "hostname")
		hostnameContent := res.onionAddr + ".onion\n"
		if err := os.WriteFile(hostnamePath, []byte(hostnameContent), 0600); err != nil {
			log.Fatalf("Failed to write hostname: %v", err)
		}
		log.Printf("Hostname saved to: %s", hostnamePath)
	} else {
		log.Printf("Vanity Onion Address: %s", res.onionAddr)
		log.Printf("Public Key (hex): %s", hex.EncodeToString(res.finalPub))
		log.Printf("Expanded Private Key (hex): %s", hex.EncodeToString(res.expanded[:]))
		log.Printf("Total Attempts: %d", res.attempts)
	}
}
