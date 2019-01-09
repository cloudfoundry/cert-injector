package main

import (
	"certificate-injector/command"
	"certificate-injector/container"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"code.cloudfoundry.org/hydrator/oci-directory"
)

const (
	LockFileName    = "GrootRootfsMutex"
	grootBin        = "c:\\var\\vcap\\packages\\groot\\groot.exe"
	wincBin         = "c:\\var\\vcap\\packages\\winc\\winc.exe"
	diffExporterBin = "c:\\var\\vcap\\packages\\diff-exporter\\diff-exporter.exe"
	hydrateBin      = "c:\\var\\vcap\\packages\\hydrate\\hydrate.exe"
)

type image interface {
	ContainsHydratorAnnotation(ociImagePath string) (bool, error)
}

type cmd interface {
	Run(executable string, args ...string) error
}

type conf interface {
	Write(certData []byte) error
}

func Run(args []string, image image, cmd cmd, conf conf) error {
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

	// TODO: for each image_uri, check if it contains an annotation, remove that layer
	// TODO: parse the arg image_uri for the oci image path
	ociImageUri := args[3]
	err = cmd.Run(hydrateBin, "remove-layer", "-ociImage", ociImageUri)
	if err != nil {
		return fmt.Errorf("hydrate.exe remove-layer failed: %s\n", err)
	}

	err = conf.Write(certData)
	if err != nil {
		return fmt.Errorf("Write container config failed: %s", err)
	}

	/*grootDriverStore := args[1]
	grootImageUris := args[2:]

	// TODO: the hydrator that is altering the rootfs is not looking at this lock file anyway, so we are punting right now
	lock, err := filelock.NewLocker(filepath.Join(os.TempDir(), LockFileName)).Open()
	if err != nil {
		return fmt.Errorf("open lock: %s\n", err)
	}
	defer lock.Close()



	// workaround for https://github.com/Microsoft/hcsshim/issues/155
	fmt.Printf("%s\n", "Deleting existing containers")
	err = exec.Command(fmt.Sprintf("Get-ComputeProcess | foreach { & %s delete $_.Id }", wincBin)).Run()
	if err != nil {
		return fmt.Errorf("Cannot delete existing containers\n")
	}

	files, err := ioutil.ReadDir(fmt.Sprintf("%s\\volumes", grootDriverStore))
	if !os.IsNotExist(err) {
		return fmt.Errorf("groot delete failed: %s\n", err)
	}

	for _, file := range files {
		err = exec.Command(fmt.Sprintf("%s --driver-store %s delete %s", grootBin, grootDriverStore, file.Name())).Run()
		if err != nil {
			return fmt.Errorf("groot delete failed: %s\n", err)
		}
	}



	fmt.Printf("%s\n", "Begin exporting layer")
	for _, uri := range grootImageUris {
		containerId := fmt.Sprintf("layer%d", int32(time.Now().Unix()))

		fmt.Printf("%s\n", "Creating Volume")
		cmd := exec.Command(fmt.Sprintf("%s --driver-store %s create %s", grootBin, grootDriverStore, uri))
		var stdoutBuffer bytes.Buffer
		cmd.Stdout = &stdoutBuffer
		cmd.Run()
		if err != nil {
			return fmt.Errorf("Groot create failed\n")
		}

		var config map[string]interface{}
		if err := json.Unmarshal(stdoutBuffer.Bytes(), &config); err != nil {
			return fmt.Errorf("failed to parse process spec\n")
		}


		fmt.Printf("Writing config.json")
		bundleDir := filepath.Join(os.TempDir(), containerId)
		configPath := filepath.Join(bundleDir, "config.json")
		if err = os.Mkdir(bundleDir, 0755); err != nil {
			return fmt.Errorf("Failed to create bundle directory\n")
		}

		configBytes, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("Failed to write config.json\n")
		}
		configFile, err := os.Create(configPath)
		if err != nil {
			return fmt.Errorf("Failed to create config.json\n")
		}
		defer configFile.Close()

		_, err = configFile.Write(configBytes)
		if err != nil {
			return fmt.Errorf("Failed to write config.json\n")
		}

		configFile.Sync()

		fmt.Printf("%s\n", "winc run")
		err = exec.Command(fmt.Sprintf("%s run -b %s %s", wincBin, bundleDir, containerId)).Run()
		if err != nil {
			return fmt.Errorf("winc run failed\n")
		}

		fmt.Printf("%s\n", "Running diff-exporter")
		diffOutputFile := filepath.Join(os.TempDir(), fmt.Sprintf("diff-output%d", int32(time.Now().Unix())))
		err = exec.Command(fmt.Sprintf("%s -outputFile %s -containerId %s -bundlePath %s", diffExporterBin, diffOutputFile, containerId, bundleDir)).Run()
		if err != nil {
			return fmt.Errorf("diff-exporter failed\n")
		}

		fmt.Printf("%s\n", "Running hydrator")
		ociImage := strings.Split(uri, "///")[1]
		ociImage = strings.Replace(ociImage, "/", "\\", -1)
		err = exec.Command(fmt.Sprintf("%s add-layer -ociImage %s -layer %s", hydrateBin, ociImage, diffOutputFile)).Run()
		if err != nil {
			return fmt.Errorf("hydrator failed\n")
		}

		fmt.Printf("%s\n", "Cleaning up")
		err = exec.Command(fmt.Sprintf("%s --driver-store %s delete %s", grootBin, grootDriverStore, containerId)).Run()
		if err != nil {
			return fmt.Errorf("groot delete failed\n")
		}
		err = os.RemoveAll(diffOutputFile)
		if err != nil {
			return fmt.Errorf("diff output file deletion failed\n")
		}
	}
	*/
	return nil
}

func main() {

	logger := log.New(os.Stderr, "", 0)
	handler := directory.NewHandler("ociImageDir")
	image := container.NewImage(handler)
	cmd := command.NewCmd()
	conf := container.NewConfig()
	if err := Run(os.Args, image, cmd, conf); err != nil {
		logger.Print(err)
		os.Exit(1)
	}
}
