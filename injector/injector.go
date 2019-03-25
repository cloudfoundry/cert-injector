package injector

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	grootBin        = "c:\\var\\vcap\\packages\\groot\\groot.exe"
	wincBin         = "c:\\var\\vcap\\packages\\winc\\winc.exe"
	diffExporterBin = "c:\\var\\vcap\\packages\\diff-exporter\\diff-exporter.exe"
	hydrateBin      = "c:\\var\\vcap\\packages\\hydrate\\hydrate.exe"
)

type cmd interface {
	Run(executable string, args ...string) (string, string, error)
}

type config interface {
	Write(bundleDir, grootOutput, certData string) error
}

type logger interface {
	Println(v ...interface{})
}

type Injector struct {
	cmd    cmd
	config config
	stdout logger
	stderr logger
}

func NewInjector(cmd cmd, config config, stdout, stderr logger) Injector {
	return Injector{
		cmd:    cmd,
		config: config,
		stdout: stdout,
		stderr: stderr,
	}
}

func (i Injector) InjectCert(grootDriverStore, uri, certDirectory string) error {
	_, _, err := i.cmd.Run(hydrateBin, "remove-layer", "-ociImage", uri)
	if err != nil {
		return fmt.Errorf("hydrate remove-layer -ociImage %s failed: %s\n", uri, err)
	}

	containerId := fmt.Sprintf("layer-%d", int32(time.Now().Unix()))

	grootOutput, stderr, err := i.cmd.Run(grootBin, "--driver-store", grootDriverStore, "create", uri, containerId)
	if err != nil {
		i.stdout.Println(grootOutput)
		i.stderr.Println(stderr)
		return fmt.Errorf("groot create failed: %s", err)
	}
	defer func() {
		stdout, stderr, err := i.cmd.Run(grootBin, "--driver-store", grootDriverStore, "delete", containerId)
		if err != nil {
			i.stdout.Println("groot delete failed")
			i.stdout.Println(stdout)
			i.stderr.Println(stderr)
		}
	}()

	bundleDir := filepath.Join(os.TempDir(), containerId)
	err = os.MkdirAll(bundleDir, 0755)
	if err != nil {
		return fmt.Errorf("create bundle directory failed: %s", err)
	}
	defer os.RemoveAll(bundleDir)

	err = i.config.Write(bundleDir, grootOutput, certDirectory)
	if err != nil {
		return fmt.Errorf("container config write failed: %s", err)
	}

	stdout, stderr, err := i.cmd.Run(wincBin, "run", "-b", bundleDir, containerId)
	if err != nil {
		i.stdout.Println(stdout)
		i.stderr.Println(stderr)
		return fmt.Errorf("winc run failed: %s", err)
	}

	diffOutputFile := filepath.Join(os.TempDir(), fmt.Sprintf("diff-output%d", int32(time.Now().Unix())))
	_, _, err = i.cmd.Run(diffExporterBin, "-outputFile", diffOutputFile, "-containerId", containerId, "-bundlePath", bundleDir)
	if err != nil {
		return fmt.Errorf("diff-exporter failed exporting the layer: %s", err)
	}
	defer os.RemoveAll(diffOutputFile)

	_, _, err = i.cmd.Run(hydrateBin, "add-layer", "-ociImage", uri, "-layer", diffOutputFile)
	if err != nil {
		return fmt.Errorf("hydrate add-layer failed: %s", err)
	}

	return nil
}
