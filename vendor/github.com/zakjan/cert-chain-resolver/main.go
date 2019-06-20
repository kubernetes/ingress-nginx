package main

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/zakjan/cert-chain-resolver/certUtil"
	"io/ioutil"
	"os"
)

func openInputFile(filename string) (*os.File, error) {
	if filename == "" {
		return os.Stdin, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func openOutputFile(filename string) (*os.File, error) {
	if filename == "" {
		return os.Stdout, nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func run(inputFilename string, outputFilename string, outputIntermediateOnly bool,
	outputDerFormat bool, includeSystem bool) error {
	inputFile, err := openInputFile(inputFilename)
	if err != nil {
		return err
	}

	outputFile, err := openOutputFile(outputFilename)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(inputFile)
	if err != nil {
		return err
	}

	cert, err := certUtil.DecodeCertificate(data)
	if err != nil {
		return err
	}

	certs, err := certUtil.FetchCertificateChain(cert)
	if err != nil {
		return err
	}

	if includeSystem {
		certs, err = certUtil.AddRootCA(certs)
		if err != nil {
			return err
		}
	}

	if outputIntermediateOnly {
		certs = certs[1:]
	}

	if !outputDerFormat {
		data = certUtil.EncodeCertificates(certs)
	} else {
		data = certUtil.EncodeCertificatesDER(certs)
	}

	_, err = outputFile.Write(data)
	if err != nil {
		return err
	}

	for i, cert := range certs {
		fmt.Fprintf(os.Stderr, "%d: %s\n", i+1, cert.Subject.CommonName)
	}
	fmt.Fprintf(os.Stderr, "Certificate chain complete.\n")
	fmt.Fprintf(os.Stderr, "Total %d certificate(s) found.\n", len(certs))

	return nil
}

func main() {
	var (
		inputFilename          string
		outputFilename         string
		outputIntermediateOnly bool
		outputDerFormat        bool
		includeSystem          bool
	)

	app := cli.NewApp()
	app.Usage = "SSL certificate chain resolver"
	app.ArgsUsage = "[INPUT_FILE]"
	app.Version = "1.0.1"
	app.HideHelp = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "output, o",
			Usage:       "output to `OUTPUT_FILE` (default: stdout)",
			Destination: &outputFilename,
		},
		cli.BoolFlag{
			Name:        "intermediate-only, i",
			Usage:       "output intermediate certificates only",
			Destination: &outputIntermediateOnly,
		},
		cli.BoolFlag{
			Name:        "der, d",
			Usage:       "output DER format",
			Destination: &outputDerFormat,
		},
		cli.BoolFlag{
			Name:        "include-system, s",
			Usage:       "include root CA from system in output",
			Destination: &includeSystem,
		},
	}
	app.Action = func(c *cli.Context) error {
		args := c.Args()
		if len(args) > 0 {
			inputFilename = args[0]
		}

		err := run(inputFilename, outputFilename, outputIntermediateOnly, outputDerFormat, includeSystem)
		return err
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
