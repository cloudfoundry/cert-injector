package main

import (
	"log"
	"os"

	"code.cloudfoundry.org/cert-injector/command"
	"code.cloudfoundry.org/cert-injector/container"
	"code.cloudfoundry.org/cert-injector/injector"
)

func main() {
	args := os.Args

	stdout := log.New(os.Stdout, "", 0)
	stderr := log.New(os.Stderr, "", 0)

	cmd := command.NewCmd()
	config := container.NewConfig()

	inj := injector.NewInjector(cmd, config, stdout, stderr)

	// There can be multiple image uris because groot.cached_image_uris is an array.
	if len(args) < 4 {
		log.Fatalf("usage: %s <driver_store> <cert_directory> <image_uri>...\n", args[0])
	}

	driverStore := args[1]
	certDirectory := args[2]
	ociImageUris := args[3:]

	for _, uri := range ociImageUris {
		err := inj.InjectCert(driverStore, uri, certDirectory)
		if err != nil {
			log.Fatalf("cert-injector failed: %s", err)
		}
	}
}
