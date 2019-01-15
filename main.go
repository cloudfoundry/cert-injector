package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cert-injector/command"
	"code.cloudfoundry.org/cert-injector/container"
)

const (
	grootBin        = "c:\\var\\vcap\\packages\\groot\\groot.exe"
	wincBin         = "c:\\var\\vcap\\packages\\winc\\winc.exe"
	diffExporterBin = "c:\\var\\vcap\\packages\\diff-exporter\\diff-exporter.exe"
	hydrateBin      = "c:\\var\\vcap\\packages\\hydrate\\hydrate.exe"
)

type cmd interface {
	Run(executable string, args ...string) ([]byte, []byte, error)
}

type conf interface {
	Write(bundleDir string, grootOutput []byte, certData []byte) error
}

func Run(args []string, cmd cmd, conf conf) error {
	// There are multiple image uris because groot.cached_image_uris is an array.
	if len(args) < 4 {
		return fmt.Errorf("usage: %s <driver_store> <cert_file> <image_uri>...\n", args[0])
	}

	certFile := args[2]
	certData, err := ioutil.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("Failed to read cert_file: %s", err)
	}

	if len(certData) == 0 {
		return nil
	}

	ociImageUris := args[3:]

	for _, uri := range ociImageUris {
		_, _, err = cmd.Run(hydrateBin, "remove-layer", "-ociImage", uri)
		if err != nil {
			return fmt.Errorf("hydrate.exe remove-layer -ociImage %s failed: %s\n", uri, err)
		}
	}

	grootDriverStore := args[1]

	for _, uri := range ociImageUris {
		grootOutput, stderr, err := cmd.Run(grootBin, "--driver-store", grootDriverStore, "create", uri)
		if err != nil {
			os.Stdout.Write(grootOutput)
			os.Stderr.Write(stderr)
			return fmt.Errorf("groot create failed: %s", err)
		}

		containerId := fmt.Sprintf("layer-%d", int32(time.Now().Unix()))
		bundleDir := filepath.Join(os.TempDir(), containerId)
		err = os.MkdirAll(bundleDir, 0755)
		if err != nil {
			return fmt.Errorf("Failed to create bundle directory: %s\n", err)
		}

		err = conf.Write(bundleDir, grootOutput, certData)
		if err != nil {
			return fmt.Errorf("Write container config failed: %s", err)
		}

		_, _, err = cmd.Run(wincBin, "run", "-b", bundleDir, containerId)
		if err != nil {
			return fmt.Errorf("winc run failed: %s", err)
		}

		diffOutputFile := filepath.Join(os.TempDir(), fmt.Sprintf("diff-output%d", int32(time.Now().Unix())))
		_, _, err = cmd.Run(diffExporterBin, "-outputFile", diffOutputFile, "-containerId", containerId, "-bundlePath", bundleDir)
		if err != nil {
			return fmt.Errorf("diff-exporter failed exporting the layer: %s", err)
		}

		_, _, err = cmd.Run(hydrateBin, "add-layer", "-ociImage", uri, "-layer", diffOutputFile)
		if err != nil {
			return fmt.Errorf("hydrate add-layer failed: %s", err)
		}

		_, _, err = cmd.Run(grootBin, "--driver-store", grootDriverStore, "delete", uri)
		if err != nil {
			return fmt.Errorf("groot delete failed: %s", err)
		}

		err = os.RemoveAll(bundleDir)
		if err != nil {
			return fmt.Errorf("remove bundle directory failed: %s", err)
		}
	}

	return nil
}

func main() {
	cmd := command.NewCmd()
	conf := container.NewConfig()

	err := Run(os.Args, cmd, conf)
	if err != nil {
		log.Fatalf("cert-injector failed: %s", err)
	}
}
