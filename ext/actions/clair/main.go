// Package klarscan uses klar's codebase to scan docker images
package klarscan

import (
	"time"
	"fmt"

	"github.com/ObjectifLibre/csf/actions"

	"gopkg.in/yaml.v2"
	"github.com/optiopay/klar/clair"
	"github.com/optiopay/klar/docker"
)

func init() {
	actions.RegisterActionModule("klar", &clairActionModule{})
}

var _ actions.ActionModuleInterface = clairActionModule{}

var config klarConfig

type clairActionModule struct {}

type dockerConfig struct {
	User             string `yaml:"user"`
	Password         string `yaml:"password"`
	Token            string `yaml:"token"`
	InsecureTLS      bool `yaml:"insecureTLS"`
	InsecureRegistry bool `yaml:"insecureRegistry"`
	Timeout          int `yaml:"timeout"`
}

type klarConfig struct {
	ClairAddr     string `yaml:"clairAddress"`
	ClairOutput   string `yaml:"clairOutput"`
	Threshold     int `yaml:"threshold"`
	ClairTimeout  int `yaml:"clairTimeout"`
	DockerConfig  dockerConfig `yaml:"dockerConfig"`
}

// Actions exposes 2 actions, scan_image to scan one image and scan_images to
// scan multiple images at once.
func (clair clairActionModule) Actions() (map[string][]actions.ArgType, map[string][]actions.ArgType) {
	in := map[string][]actions.ArgType{
		"scan_image": {{T: "string, docker image name", N: "image"}},
		"scan_images": {{T: "[]string, array of docker image names", N: "images"},
		}}
	out := map[string][]actions.ArgType{
		"scan_image": {{T: "array of vulnerabilities", N: "vulns"}},
		"scan_images": {{T: "array of vulnerabilities", N: "{{name of the image scanned}}"},
		}}
	return in, out
}

func getDockerConfig(image string) (docker.Config)  {
	dockercfg := docker.Config{
		ImageName: image,
		User: config.DockerConfig.User,
		Password: config.DockerConfig.Password,
		Token: config.DockerConfig.Token,
		InsecureTLS: config.DockerConfig.InsecureTLS,
		InsecureRegistry: config.DockerConfig.InsecureRegistry,
		Timeout: time.Duration(config.DockerConfig.Timeout) * time.Second,
	}
	return dockercfg
}

func scanDockerImage(data map[string]interface{}) (map[string]interface{}, error)  {
	imagename, ok := data["image"].(string)
	if !ok {
		return nil, fmt.Errorf("Expected 'image' as string")
	}
	if vulns, err := scanSingleImage(imagename); err != nil {
		return nil, err
	} else {
		result := make(map[string]interface{})
		result["vulns"] = vulns
		return result, nil
	}
}

func scanDockerImages(data map[string]interface{}) (map[string]interface{}, error) {
	images, ok := data["images"].([]string)
	if !ok {
		return nil, fmt.Errorf("Expected 'images' as string array")
	}
	result := make(map[string]interface{})
	errs := make(chan error)
	for _, image := range(images) {
		go func(image string) {
			vulns, err := scanSingleImage(image)
			if err != nil {
				errs <- err
				return
			}
			result[image] = vulns
			errs <- nil
		}(image)
	}
	for i := 0; i < len(images); i += 1 {
		err := <- errs
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func scanSingleImage(imagename string) (interface{}, error) {
	curdockerconfig := getDockerConfig(imagename)
	image, err := docker.NewImage(&curdockerconfig)
	if err != nil {
		return nil, fmt.Errorf("Could not get docker image: %s", err)
	}
	err = image.Pull()
	if err != nil {
		return nil, err
	}
	if len(image.FsLayers) < 1 {
		return nil, fmt.Errorf("Could not pull fslayers: image has no fslayers")
	}
	c := clair.NewClair(config.ClairAddr, 3, time.Duration(config.ClairTimeout) * time.Second)
	var vs []*clair.Vulnerability
	vs, err = c.Analyse(image)
	if err != nil {
		return nil, fmt.Errorf("Could not analyse image: %s", err)
	}
	vulns := make([]clair.Vulnerability, len(vs))
	for i, vuln := range(vs) {
		vulns[i] = *vuln
	}
	return vulns, nil
}

func clairActionHandler(action string, data map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "scan_image":
		return scanDockerImage(data)
	case "scan_images":
		return scanDockerImages(data)
	default:
		return nil, fmt.Errorf("No action named %s", action)
	}
}

func (clair clairActionModule) Setup(rawcfg []byte) error {
	if err := yaml.Unmarshal(rawcfg, &config); err != nil {
		return fmt.Errorf("Error in config file: %s", err)
	}
	return nil
}

func (clair clairActionModule) ActionHandler() func(string, map[string]interface{}) (map[string]interface{}, error) {
	return clairActionHandler
}
