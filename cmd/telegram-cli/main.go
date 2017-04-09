package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/andreyvit/telegramapi"
	"github.com/andreyvit/telegramapi/mtproto"
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
	var err error

	options := telegramapi.Options{
		Endpoint:  productionEndpoint,
		PublicKey: publicKey,
		Verbose:   2,
	}

	apiID := os.Getenv("TG_APP_ID")
	if apiID == "" {
		fmt.Fprintf(os.Stderr, "** missing TG_APP_ID env variable\n")
		os.Exit(64) // EX_USAGE
	}
	options.APIID, err = strconv.Atoi(apiID)
	if apiID == "" {
		fmt.Fprintf(os.Stderr, "** invalid TG_APP_ID\n")
		os.Exit(64) // EX_USAGE
	}

	options.APIHash = os.Getenv("TG_API_HASH")
	if options.APIHash == "" {
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

	go run(conn)

	err = doit(conn)
	if err != nil {
		log.Printf("** ERROR: process: %v", err)
	}

	log.Printf("✓ DONE")
}

func run(tg *telegramapi.Conn) {
	tg.Run()
	err := tg.Err()
	if err != nil {
		log.Printf("** ERROR: session: %v", err)
		os.Exit(1)
	}
}

func doit(tg *telegramapi.Conn) error {
	tg.Session.WaitReady()

	r, err := tg.Session.Send(&mtproto.TLHelpGetNearestDC{})
	if err != nil {
		return err
	}

	r, err = tg.Session.Send(&mtproto.TLAuthSendCode{
		Flags:         1,
		PhoneNumber:   "79061932959",
		CurrentNumber: true,
		APIID:         tg.APIID,
		APIHash:       tg.APIHash,
	})
	if err != nil {
		return err
	}

	log.Printf("Got auth.sendCode response: %v", r)

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
