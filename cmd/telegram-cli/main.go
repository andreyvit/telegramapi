package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/andreyvit/telegramapi/tl"
	"github.com/chzyer/readline"
	"github.com/kr/pretty"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/andreyvit/telegramapi"
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

	var isTest bool
	var isFreshStart bool
	var dumpStateAndQuit bool
	flag.StringVar(&tool.phoneNumber, "phone", "", "Set the phone number to log in as")
	flag.BoolVar(&isTest, "test", false, "Use test endpoint")
	flag.BoolVar(&isFreshStart, "F", false, "Kill state and start any")
	flag.BoolVar(&dumpStateAndQuit, "dump", false, "Dump state and quit")
	flag.Parse()

	if isTest {
		options.SeedAddr = telegramapi.Addr{"149.154.167.40", 443}
	}

	state := new(telegramapi.State)
	stateBytes, err := ioutil.ReadFile(tool.stateFile)
	if err == nil && !isFreshStart {
		err = tl.ReadBare(state, stateBytes)
		if err != nil {
			log.Printf("** ERROR: reading state from %v: %v", tool.stateFile, err)
			os.Exit(1)
		}

		if state.DCs[2] != nil && state.DCs[2].Auth.KeyID == 0 {
			log.Printf("** oops: %v", pretty.Sprint(state))
			os.Exit(1)
		}
	}

	if dumpStateAndQuit {
		log.Printf("State: %v", pretty.Sprint(state))
		os.Exit(0)
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

	phoneNumber string
	phoneCode   string

	authed bool
}

func (tool *Tool) HandleConnectionReady() {
	go tool.runProcessingNoErr()
}
func (tool *Tool) HandleStateChanged(newState *telegramapi.State) {
	log.Printf("HandleStateChanged: %v", pretty.Sprint(newState))
	if tool.authed && newState.DCs[2] != nil && newState.DCs[2].Auth.KeyID == 0 {
		panic("** oops")
		// panic(fmt.Sprint("** oops: %v", pretty.Sprint(newState)))
	}
	if newState.DCs[2] != nil && newState.DCs[2].Auth.KeyID != 0 {
		tool.authed = true
	}

	bytes := tl.BareBytes(newState)
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
	if tool.tg.LoginState() == telegramapi.LoggedOut {
		err := tool.tg.StartLogin(tool.phoneNumber)
		if err != nil {
			return err
		}
	}

	if tool.tg.LoginState() == telegramapi.WaitingForCode {
		code, err := readline.Line("Code: ")
		if err != nil {
			return err
		}

		err = tool.tg.CompleteLoginWithCode(code)
		if err != nil {
			return err
		}
	}

	if tool.tg.LoginState() == telegramapi.WaitingFor2FA {
		pw, err := readline.Password("2FA password: ")
		if err != nil {
			return err
		}

		err = tool.tg.CompleteLoginWith2FAPassword(pw)
		if err != nil {
			return err
		}
	}

	if tool.tg.LoginState() != telegramapi.LoggedIn {
		log.Printf("** login failed for some reason")
		return errors.New("login failed for some reason")
	}

	log.Printf("LOGGED IN")

	contacts := telegramapi.NewContactList()

	err := tool.tg.LoadChats(contacts)
	if err != nil {
		return err
	}

	log.Printf("Loaded contacts: %v", pretty.Sprint(contacts))

	if contacts.SelfChat != nil {
		err := tool.tg.LoadHistory(contacts, contacts.SelfChat)
		if err != nil {
			return err
		}
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
