package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/andreyvit/telegramapi"
)

const testEndpoint = "149.154.167.40:443"
const productionEndpoint = "149.154.167.50:443"

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
		Endpoint:  testEndpoint,
		PublicKey: publicKey,
	}

	err := doit(options)
	if err != nil {
		log.Printf("** ERROR: %v", err)
		os.Exit(1)
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

	log.Printf("OK")
}

func doit(options telegramapi.Options) error {
	conn, err := telegramapi.Connect(options)
	if err != nil {
		return errors.Wrap(err, "connection")
	}
	defer conn.Close()

	log.Printf("Connected")

	err = conn.SayHello()
	if err != nil {
		return errors.Wrap(err, "saying hello")
	}

	log.Printf("Sent message")

	for {
		msg, err := conn.ReadMessage(2 * time.Second)
		if err != nil {
			return errors.Wrap(err, "read")
		}
		if msg == nil {
			break
		}
		conn.PrintMessage(msg)
	}

	return nil
}
