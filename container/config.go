package container

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	oci "github.com/opencontainers/runtime-spec/specs-go"
)

const ImportCertificatePs = `
$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }
$certFile=[System.IO.Path]::GetTempFileName()
$decodedCertData = [Convert]::FromBase64String("%s")
[IO.File]::WriteAllBytes($certFile, $decodedCertData)
Import-Certificate -CertStoreLocation Cert:\\LocalMachine\Root -FilePath $certFile
Remove-Item $certFile
`

type Config struct{}

func NewConfig() Config {
	return Config{}
}

// Write creates a file that contains the output
// of groot and a powershell script that will import
// the user-provided certificates to be run on the
// container.
func (c Config) Write(bundleDir string, grootOutput []byte, certData []byte) error {
	config := oci.Spec{}

	err := json.Unmarshal(grootOutput, &config)
	if err != nil {
		return fmt.Errorf("json unmarshal groot output: %s", err)
	}

	command := fmt.Sprintf(ImportCertificatePs, string(certData))

	encodedCommand := base64.StdEncoding.EncodeToString([]byte(command))

	config.Process = &oci.Process{
		Args: []string{"powershell.exe", "-EncodedCommand", encodedCommand},
		Cwd:  `C:\`,
	}

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
