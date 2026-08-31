package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-utility-trial/internal/trial"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "packet":
		err = packetCommand(os.Args[2:])
	case "validate":
		err = validateCommand(os.Args[2:])
	case "record":
		err = recordCommand(os.Args[2:])
	case "report":
		err = reportCommand(os.Args[2:])
	case "help", "-h", "--help":
		usage()
		return
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("gooo-utility-trial packet -output CALLER_OWNED_EMPTY_DIR")
	fmt.Println("gooo-utility-trial validate -packet PACKET_DIR -receipt RECEIPT_JSON")
	fmt.Println("gooo-utility-trial record -packet PACKET_DIR -receipt RECEIPT_JSON")
	fmt.Println("gooo-utility-trial report -packet PACKET_DIR")
}

func packetCommand(args []string) error {
	flags := flag.NewFlagSet("packet", flag.ContinueOnError)
	output := flags.String("output", "", "caller-owned empty output directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	root, err := trial.GeneratePacket(*output)
	if err != nil {
		return err
	}
	return writeJSON(map[string]any{"state": "CLOSED", "protocol_ready": "CLOSED", "utility": "UNKNOWN", "output": filepath.Clean(root)})
}

func validateCommand(args []string) error {
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	packet := flags.String("packet", "", "trial packet directory")
	receiptPath := flags.String("receipt", "", "session receipt JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	receipt, err := trial.ReadReceipt(*receiptPath)
	if err != nil {
		return err
	}
	// RecordReceipt is deliberately not used here: validate is read-only and
	// cannot turn a receipt into evidence by itself.
	result, err := trial.ValidateReceipt(*packet, receipt)
	if err != nil {
		return err
	}
	return writeJSON(result)
}

func recordCommand(args []string) error {
	flags := flag.NewFlagSet("record", flag.ContinueOnError)
	packet := flags.String("packet", "", "trial packet directory")
	receiptPath := flags.String("receipt", "", "session receipt JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	result, err := trial.RecordReceipt(*packet, *receiptPath)
	if err != nil {
		return err
	}
	return writeJSON(result)
}

func reportCommand(args []string) error {
	flags := flag.NewFlagSet("report", flag.ContinueOnError)
	packet := flags.String("packet", "", "trial packet directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if err := trial.WriteReport(*packet); err != nil {
		return err
	}
	return writeJSON(map[string]any{"state": "CLOSED", "report": filepath.Join(*packet, "utility-report.md")})
}

func writeJSON(value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Println(string(data))
	return err
}

