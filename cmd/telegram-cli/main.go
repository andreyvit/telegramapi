package main

import (
	"fmt"
	"log"
	"os"

	"github.com/andreyvit/telegramapi"
)

const testEndpoint = "149.154.167.40:443"

// const productionEndpoint = "149.154.167.50:443"
const productionEndpoint = "149.154.175.100:443"

const publicKey = `
-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAwVACPi9w23mF3tBkdZz+zwrzKOaaQdr01vAbU4E1pvkfj4sqDsm6
lyDONS789sVoD/xCS9Y0hkkC3gtL1tSfTlgCMOOul9lcixlEKzwKENj1Yz/s7daS
an9tqw3bfUV/nqgbhGX81v/+7RFAEd+RwFnK7a+XYl9sluzHRyVVaTTveB2GazTw
Efzk2DWgkBluml8OREmvfraX3bkHZJTKX4EQSjBbbdJ2ZXIsRrYOXfaA+xayEGB+
8hdlLmAjbCVfaigxX0CDqWeR1yFL9kwd9P0NsZRPsmoqVwMbMu7mStFai6aIhc3n
Slv8kg9qv1m6XHVQY3PnEw+QQtqSIXklHwIDAQAB
-----END RSA PUBLIC KEY-----
`

func main() {
	options := telegramapi.Options{
		Endpoint:  productionEndpoint,
		PublicKey: publicKey,
		Verbose:   2,
	}

	appID := os.Getenv("TG_APP_ID")
	if appID == "" {
		fmt.Fprintf(os.Stderr, "** missing TG_APP_ID env variable\n")
		os.Exit(64) // EX_USAGE
	}

	apiHash := os.Getenv("TG_API_HASH")
	if apiHash == "" {
		fmt.Fprintf(os.Stderr, "** missing TG_API_HASH env variable\n")
		os.Exit(64) // EX_USAGE
	}

	conn, err := telegramapi.Connect(options)
	if err != nil {
		log.Printf("** ERROR connecting: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	log.Printf("✓ Connected")

	conn.Run()
	err = conn.Err()
	if err != nil {
		log.Printf("** ERROR: %v", err)
		os.Exit(1)
	}

	log.Printf("✓ DONE")
}

func doit(options telegramapi.Options) error {

	// for {
	// 	msg, err := conn.ReadMessage(2 * time.Second)
	// 	if err != nil {
	// 		return errors.Wrap(err, "read")
	// 	}
	// 	if msg == nil {
	// 		break
	// 	}
	// 	// conn.PrintMessage(msg)
	// }

	return nil
}
