package main

import (
	"fmt"
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

type logger interface {
	Println(v ...interface{})
}

func Run(args []string, cmd cmd, conf conf, stdout, stderr logger) error {
	// There are multiple image uris because groot.cached_image_uris is an array.
	if len(args) < 4 {
		return fmt.Errorf("usage: %s <driver_store> <cert_data> <image_uri>...\n", args[0])
	}

	grootDriverStore := args[1]
	certData := args[2]
	ociImageUris := args[3:]

	for _, uri := range ociImageUris {
		_, _, err := cmd.Run(hydrateBin, "remove-layer", "-ociImage", uri)
		if err != nil {
			return fmt.Errorf("hydrate.exe remove-layer -ociImage %s failed: %s\n", uri, err)
		}
	}

	for _, uri := range ociImageUris {
		if err := injectCert(grootDriverStore, uri, certData, cmd, conf, stdout, stderr); err != nil {
			return err
		}
	}

	return nil
}

func injectCert(grootDriverStore, uri, certData string, cmd cmd, conf conf, stdout, stderr logger) error {
	containerId := fmt.Sprintf("layer-%d", int32(time.Now().Unix()))
	grootOutput, se, err := cmd.Run(grootBin, "--driver-store", grootDriverStore, "create", uri, containerId)
	if err != nil {
		stdout.Println(grootOutput)
		stderr.Println(se)
		return fmt.Errorf("groot create failed: %s", err)
	}
	defer func() {
		so, se, err := cmd.Run(grootBin, "--driver-store", grootDriverStore, "delete", containerId)
		if err != nil {
			stdout.Println("groot delete failed")
			stdout.Println(so)
			stderr.Println(se)
		}
	}()

	bundleDir := filepath.Join(os.TempDir(), containerId)
	err = os.MkdirAll(bundleDir, 0755)
	if err != nil {
		return fmt.Errorf("create bundle directory failed: %s", err)
	}
	defer os.RemoveAll(bundleDir)

	err = conf.Write(bundleDir, grootOutput, certData)
	if err != nil {
		return fmt.Errorf("container config write failed: %s", err)
	}

	so, se, err := cmd.Run(wincBin, "run", "-b", bundleDir, containerId)
	if err != nil {
		stdout.Println(so)
		stderr.Println(se)
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
	stdout := log.New(os.Stdout, "", 0)
	stderr := log.New(os.Stderr, "", 0)

	cmd := command.NewCmd()
	config := container.NewConfig()

	err := Run(os.Args, cmd, config, stdout, stderr)
	if err != nil {
		log.Fatalf("cert-injector failed: %s", err)
	}
}
