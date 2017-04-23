package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/andreyvit/telegramapi"
	"github.com/andreyvit/telegramapi/mtproto"
)

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
		SeedAddr:  telegramapi.Addr{"149.154.175.100", 443},
		PublicKey: publicKey,
		Verbose:   2,
	}

	isTest := false
	if isTest {
		options.SeedAddr = telegramapi.Addr{"149.154.167.40", 443}
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

	tool := &Tool{
		stateFile: "tg-state.bin",
	}

	state := new(telegramapi.State)
	stateBytes, err := ioutil.ReadFile(tool.stateFile)
	if err == nil {
		err = state.ReadBytes(stateBytes)
		if err != nil {
			log.Printf("** ERROR: reading state from %v: %v", tool.stateFile, err)
			os.Exit(1)
		}
	}

	tool.tg = telegramapi.New(options, state, tool)

	err = tool.tg.Run()
	if err != nil {
		log.Printf("** ERROR: session: %v", err)
		os.Exit(1)
	}

	log.Printf("✓ DONE")
}

type Tool struct {
	tg *telegramapi.Conn

	stateFile string
}

func (tool *Tool) HandleConnectionReady() {
	go tool.runProcessingNoErr()
}
func (tool *Tool) HandleStateChanged(newState telegramapi.State) {
	bytes := newState.Bytes()
	err := ioutil.WriteFile(tool.stateFile, bytes, 0777)
	if err != nil {
		log.Printf("** ERROR: saving state to %v: %v", tool.stateFile, err)
	}
}

func (tool *Tool) runProcessingNoErr() {
	err := tool.runProcessing()
	if err != nil {
		tool.tg.Fail(err)
	} else {
		tool.tg.Shutdown()
	}
}

func (tool *Tool) runProcessing() error {
	r, err := tool.tg.Send(&mtproto.TLHelpGetConfig{})
	if err != nil {
		return err
	}

	r, err = tool.tg.Send(&mtproto.TLAuthSendCode{
		Flags:         1,
		PhoneNumber:   "79061932959",
		CurrentNumber: true,
		APIID:         tool.tg.APIID,
		APIHash:       tool.tg.APIHash,
	})
	if err != nil {
		return err
	}
	switch r := r.(type) {
	case *mtproto.TLAuthSentCode:
		log.Printf("Got auth.sendCode response: %v", r)
	default:
		return tool.tg.HandleUnknownReply(r)
	}

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
