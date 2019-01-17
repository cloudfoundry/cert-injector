package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

const ImportCertificatePs = `$ErrorActionPreference = "Stop"; trap { $host.SetShouldExit(1) }; ls "c:\trusted_certs" | foreach {$file="c:\trusted_certs\"+$_.Name; Import-Certificate -CertStoreLocation Cert:\\LocalMachine\Root -FilePath $file}`

type Config struct{}

func NewConfig() Config {
	return Config{}
}

// Write creates the container runtime config.json file,
// using the output of groot for the Root.Path field and the Windows.LayerFolders field.
// The Process field contains a command that will add
// the user-provided certificates to the container.
// The certDirectory is the directory containing certificates that will be bind-mounted
// into the container
func (c Config) Write(bundleDir string, grootOutput []byte, certDirectory string) error {
	config := oci.Spec{}

	err := json.Unmarshal(grootOutput, &config)
	if err != nil {
		return fmt.Errorf("json unmarshal groot output: %s", err)
	}

	config.Process = &oci.Process{
		Args: []string{"powershell.exe", "-Command", ImportCertificatePs},
		Cwd:  `C:\`,
	}

	config.Mounts = []specs.Mount{{
		Destination: "c:\\trusted_certs",
		Source:      certDirectory,
	}}

	marshalledConfig, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("JSON marshal config failed: %s", err)
	}

	err = ioutil.WriteFile(filepath.Join(bundleDir, "config.json"), marshalledConfig, 0644)
	if err != nil {
		return fmt.Errorf("Write config.json failed: %s", err)
	}

	return nil
}
