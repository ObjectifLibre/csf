// Package vulsaction provides an action to launch a vulnerability scan using
// docker and vuls.io. Does not handle the vulnerability database fetch and
// update.
package vulsaction

import (
	"fmt"
	"path/filepath"
	"text/template"
	"time"
	"bytes"
	"archive/tar"
	"io"

	"gopkg.in/yaml.v2"
	"github.com/fsouza/go-dockerclient"
	"github.com/mitchellh/mapstructure"

	"github.com/ObjectifLibre/csf/actions"
)

func init() {
	actions.RegisterActionModule("vuls", &vulsModuleImplementation{})
	var err error
	vulsCfgTpl, err = template.New("cfg").Parse(vulsCfgTplString)
	if err != nil {
		panic(err)
	}
}

var _ actions.ActionModuleInterface = vulsModuleImplementation{}

type vulsModuleImplementation struct {}

type vulsConfig struct {
	DbPath string `yaml:"db_path"`
	LogsPath string `yaml:"logs_path"`
	SshKeyPath string `yaml:"ssh_key_path"`
	UseHostDocker bool `yaml:"use_host_docker_socket"`
	DockerEndpoint string `yaml:"docker_endpoint"`
}

type ActionArgs struct {
	Host string
	User string
	Port string
	DeepScan bool
}

var config = vulsConfig{}

var vulsCfgTplString = `
[servers]

[servers.server]
host = "{{.Host}}"
user = "{{.User}}"
port = "{{.Port}}"
keypath = "/root/.ssh/id_rsa"
`

var vulsCfgTpl *template.Template

func (mail vulsModuleImplementation) Actions() (map[string][]actions.ArgType,map[string][]actions.ArgType) {
	in := map[string][]actions.ArgType{
		"scan_instance": {{T: "string", N: "host"},
			{T: "string", N: "user"},
			{T: "string", N: "port"},
			{T: "bool", N: "deep_scan"},
		},
	}
	out := map[string][]actions.ArgType{
		"scan_instance": {
			{T: "string (json of vuls result)", N: "result"},
		},
	}
	return in, out
}

// getScanResults downloads an archive containing the results of the scan from
// the container that did the scan.
func getScanResults(args ActionArgs, dc *docker.Client, id string) (string, error) {
	var buffer bytes.Buffer
	opts := docker.DownloadFromContainerOptions{
		OutputStream: &buffer,
		Path: "/results/current/server.json",
	}

	if err := dc.DownloadFromContainer(id, opts); err != nil {
		return "", fmt.Errorf("Could not download result from container: %s", err)
	}

	var json bytes.Buffer

	tr := tar.NewReader(&buffer)

	_, err := tr.Next()
	if err != nil {
		return "", fmt.Errorf("Could not untar scan result: %s", err)
	}

	_, err = io.Copy(&json, tr)
	if err != nil {
		return "", fmt.Errorf("Could not untar scan result: %s", err)
	}

	return json.String(), nil
}

// setupConfig uploads an archive containing the vuls config of the scan to
// the container that will run the scan.
func setupConfig(args ActionArgs, dc *docker.Client, id string) error {
	var vulsconfig bytes.Buffer

	if err := vulsCfgTpl.Execute(&vulsconfig, args); err != nil {
		return fmt.Errorf("Could not upload config to container: %s", err)
	}

	var buff bytes.Buffer

	tw := tar.NewWriter(&buff)
	defer tw.Close()

	header := new(tar.Header)
	header.Name = "/config.toml"
	header.Size = int64(vulsconfig.Len())
	header.Mode = int64(0644)

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("Could tar vuls config: %s", err)
	}

	if _, err := io.Copy(tw, &vulsconfig); err != nil {
		return fmt.Errorf("Could not upload config to container: %s", err)
	}

	opts := docker.UploadToContainerOptions{
		InputStream: &buff,
		Path: "/",
		NoOverwriteDirNonDir: false,
	}
	if err := dc.UploadToContainer(id, opts); err != nil {
		return fmt.Errorf("Could not upload config to container: %s", err)
	}
	return nil
}


func scanInstance(data map[string]interface{}) (map[string]interface{}, error) {
	var args ActionArgs
	if err := mapstructure.Decode(data, &args); err != nil {
		return nil, fmt.Errorf("Bad config: %s", err)
	}

	client, err := docker.NewClient(config.DockerEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Coud not get docker client: %s", err)
	}

	var hostconfig docker.HostConfig
	if config.UseHostDocker {
		hostconfig = docker.HostConfig{
			Binds: []string{
				filepath.Join(config.DbPath) + ":/vuls",
				filepath.Join(config.LogsPath) + ":/var/log/vuls",
				config.SshKeyPath + ":/root/.ssh/id_rsa:ro",
				"/var/run/docker.sock:/var/run/docker.sock",
			},
		}
	} else {
		hostconfig = docker.HostConfig{
			Binds: []string{
				filepath.Join(config.DbPath) + ":/vuls/",
				filepath.Join(config.LogsPath) + ":/var/log/vuls",
				config.SshKeyPath + ":/root/.ssh/id_rsa:ro",
			},
		}
	}
	opt := docker.CreateContainerOptions{
		Name: "",
		Config: &docker.Config{
			Image: "vuls/vuls",
			Cmd: []string{"scan", "-config=/config.toml",
				"-ssh-native-insecure", "-results-dir=/results"},
                },
		HostConfig: &hostconfig,
	}
	container, err := client.CreateContainer(opt)
	if err != nil {
		return nil, fmt.Errorf("Could not create container: %s", err)
	}

	if err := setupConfig(args, client, container.ID); err != nil {
		return nil, err
	}

	if err = client.StartContainer(container.ID, &docker.HostConfig{}); err != nil {
		return nil, fmt.Errorf("Could not start container: %s", err)
	}
	for {
		container, err := client.InspectContainer(container.ID)
		if err != nil {
			return nil, fmt.Errorf("Could inspect container: %s", err)
		}
		if container.State.Running {
			time.Sleep(1 * time.Second)
			continue
		} else {
			break
		}
	}
	if json, err := getScanResults(args, client, container.ID); err != nil {
		return nil, err
	} else {
		result := map[string]interface{}{
			"result": json,
		}
		return result, nil
	}
}

func vulsHandler(action string, data map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "scan_instance":
 		return scanInstance(data)
	default:
		return nil, fmt.Errorf("No such action: " + action)
	}
}

func (mail vulsModuleImplementation) Setup(rawcfg []byte) error {
	if err := yaml.Unmarshal(rawcfg, &config); err != nil {
		return fmt.Errorf("Bad config: %s", err)
	}
	return nil
}

func (mail vulsModuleImplementation) ActionHandler() func(string, map[string]interface{}) (map[string]interface{}, error) {
	return vulsHandler
}
