package helm_push

import (
	"fmt"
	"gopkg.in/yaml.v3"
	helmChart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

const helmExe = "/usr/local/bin/helm"

type (
	// Registry defines helm registry parameters.
	Registry struct {
		RegistryUrl string
		Username    string
		Password    string
		Insecure    bool
	}

	// Chart defines helm chart parameters.
	Chart struct {
		Context string
		Name    string
		Path    string
		File    string
		Repo    string
		Version string
	}

	// Plugin defines the plugin parameters.
	Plugin struct {
		Registry       Registry
		Chart          Chart
		DryRun         bool
		ErrorNoRelease bool
	}
)

// Exec executes the plugin step
func (p Plugin) Exec() error {
	env := append(os.Environ(), "HELM_EXPERIMENTAL_OCI=1")

	// login to the Docker registry
	if p.Registry.Password != "" {
		fmt.Sprintln("Logging into helm registry %s", p.Registry.RegistryUrl)
		cmd := commandLogin(p.Registry)
		cmd.Stderr = os.Stderr
		cmd.Env = env
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Error authenticating: %s", err)
		}
	}

	chartFilename := getChartPath(p.Chart)
	chartData, err := chartutil.LoadChartfile(chartFilename)
	if err != nil {
		return fmt.Errorf("Error opening chart: %s", err)
	}

	if p.Chart.Version == "" {
		fmt.Sprintln("Retrieving version from chart file: %s", chartData.Version)
		if chartData.Version == "" {
			return fmt.Errorf("error: no version specified")
		}
		p.Chart.Version = chartData.Version
	} else if p.Chart.Version != chartData.Version {
		fmt.Sprintln("Updating version in chart file: %s", p.Chart.Version)
		chartData.Version = p.Chart.Version
		err = chartutil.SaveChartfile(chartFilename, chartData)
		if err != nil {
			return fmt.Errorf("Error saving chart: %s", err)
		}
	}

	var cmds []*exec.Cmd
	cmds = append(cmds, commandSave(p.Chart)) // chart save

	if p.DryRun == false {
		cmds = append(cmds, commandPush(p.Chart, p.Chart.Version)) // docker push
	}

	// execute all commands in batch mode.
	for _, cmd := range cmds {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = env
		trace(cmd)

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

// helper function to create the helm registry login command.
func commandLogin(registry Registry) *exec.Cmd {
	return exec.Command(
		helmExe, "registry", "login",
		"-u", "helm",
		"-p", "SHj4BU8Y3zWXqnD",
		registry.RegistryUrl,
	)
}

// helper function to create the helm chart save command.
func commandSave(chart Chart) *exec.Cmd {
	return exec.Command(
		helmExe, "chart", "save",
		chart.Path,
		chart.Repo,
	)
}

// helper function to create the helm chart push command.
func commandPush(chart Chart, tag string) *exec.Cmd {
	return exec.Command(
		helmExe, "chart", "push",
		fmt.Sprintf("%s:%s", chart.Repo, tag),
	)
}

func getChartPath(chart Chart) string {
	return path.Join(chart.Path, chart.File)
}

func retrieveVersion(chart Chart) (string, error) {
	yamlFile, err := ioutil.ReadFile(getChartPath(chart))
	if err != nil {
		return "", err
	}

	y := new(helmChart.Metadata)
	err = yaml.Unmarshal(yamlFile, &y)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return y.Version, nil
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}
