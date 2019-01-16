package container

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/text/encoding/unicode"
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

// Write creates the container runtime config.json file,
// using the output of groot for the Root.Path field and the Windows.LayerFolders field.
// The Process field contains a command that will add
// the user-provided certificates to the container.
func (c Config) Write(bundleDir string, grootOutput []byte, certData string) error {
	config := oci.Spec{}

	err := json.Unmarshal(grootOutput, &config)
	if err != nil {
		return fmt.Errorf("json unmarshal groot output: %s", err)
	}

	command := fmt.Sprintf(ImportCertificatePs, certData)
	// The command variable contains a UTF-8 string. However, the -EncodedCommand
	// argument to powershell expects a Base-64 encoded UTF-16 string. So we convert
	// our string to UTF-16 before Base-64 encoding it.
	encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	utf16command, err := encoder.Bytes([]byte(command))
	if err != nil {
		return fmt.Errorf("could not convert command to UTF-16 %s", err.Error())
	}

	encodedCommand := base64.StdEncoding.EncodeToString(utf16command)

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
