package main

import (
	"fmt"
	"io"
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
	Write(bundleDir string, grootOutput []byte, certData string) error
}

func Run(args []string, cmd cmd, conf conf, stdout, stderr io.Writer) error {
	// There are multiple image uris because groot.cached_image_uris is an array.
	if len(args) < 4 {
		return fmt.Errorf("usage: %s <driver_store> <cert_data> <image_uri>...\n", args[0])
	}

	certData := args[2]

	ociImageUris := args[3:]

	for _, uri := range ociImageUris {
		_, _, err := cmd.Run(hydrateBin, "remove-layer", "-ociImage", uri)
		if err != nil {
			return fmt.Errorf("hydrate.exe remove-layer -ociImage %s failed: %s\n", uri, err)
		}
	}

	grootDriverStore := args[1]

	for _, uri := range ociImageUris {
		if err := injectCert(grootDriverStore, uri, certData, cmd, conf, stdout, stderr); err != nil {
			return err
		}
	}

	return nil
}

func injectCert(grootDriverStore, uri, certData string, cmd cmd, conf conf, stdout, stderr io.Writer) error {
	containerId := fmt.Sprintf("layer-%d", int32(time.Now().Unix()))
	grootOutput, se, err := cmd.Run(grootBin, "--driver-store", grootDriverStore, "create", uri, containerId)
	if err != nil {
		// TODO: write to stdout/stderr better
		stdout.Write(grootOutput)
		stderr.Write(se)
		return fmt.Errorf("groot create failed: %s", err)
	}
	defer func() {
		so, se, err := cmd.Run(grootBin, "--driver-store", grootDriverStore, "delete", containerId)
		if err != nil {
			stdout.Write([]byte("groot delete failed\n"))
			stdout.Write(so)
			stderr.Write(se)
		}
	}()

	bundleDir := filepath.Join(os.TempDir(), containerId)
	err = os.MkdirAll(bundleDir, 0755)
	if err != nil {
		return fmt.Errorf("Failed to create bundle directory: %s\n", err)
	}
	defer os.RemoveAll(bundleDir)

	err = conf.Write(bundleDir, grootOutput, certData)
	if err != nil {
		return fmt.Errorf("Write container config failed: %s", err)
	}

	so, se, err := cmd.Run(wincBin, "run", "-b", bundleDir, containerId)
	if err != nil {
		// TODO: write to stdout/stderr better
		stdout.Write(so)
		stderr.Write(se)
		return fmt.Errorf("winc run failed: %s", err)
	}

	diffOutputFile := filepath.Join(os.TempDir(), fmt.Sprintf("diff-output%d", int32(time.Now().Unix())))
	_, _, err = cmd.Run(diffExporterBin, "-outputFile", diffOutputFile, "-containerId", containerId, "-bundlePath", bundleDir)
	if err != nil {
		return fmt.Errorf("diff-exporter failed exporting the layer: %s", err)
	}
	defer os.RemoveAll(diffOutputFile)

	_, _, err = cmd.Run(hydrateBin, "add-layer", "-ociImage", uri, "-layer", diffOutputFile)
	if err != nil {
		return fmt.Errorf("hydrate add-layer failed: %s", err)
	}

	return nil
}

func main() {
	cmd := command.NewCmd()
	conf := container.NewConfig()

	err := Run(os.Args, cmd, conf, os.Stdout, os.Stderr)
	if err != nil {
		log.Fatalf("cert-injector failed: %s", err)
	}
}
