package helm_push

import (
	"encoding/base64"
	"fmt"
	"github.com/smoothify/drone-helm-push/pkg/helm/chartutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type (
	// Helm defines helm repo and registry parameters.
	Helm struct {
		RegistryUrl string
		RepoUrl     string
		Username    string
		Password    string
		Insecure    bool
		Oci         bool
		Legacy      bool
	}

	// Chart defines helm chart parameters.
	Chart struct {
		Context string
		Name    string
		Path    string
		File    string
		Version string
		OciUrl  string
	}

	// Plugin defines the plugin parameters.
	Plugin struct {
		Helm           Helm
		Chart          Chart
		DryRun         bool
		ErrorNoRelease bool
	}
)

// Exec executes the plugin step
func (p Plugin) Exec() error {
	env := os.Environ()

	if p.Helm.Oci {
		p.Chart.OciUrl = getChartOciUrl(p)
		env = append(env, "HELM_EXPERIMENTAL_OCI=1")

		// login to the Helm OCI registry
		if p.Helm.Password != "" {
			fmt.Sprintln("Logging into helm oci registry %s", p.Helm.RegistryUrl)
			cmd := commandOciLogin(p.Helm)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = env
			err := cmd.Run()
			if err != nil {
				return fmt.Errorf("Error authenticating: %s", err)
			}
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

	if p.Helm.Legacy {
		env = append(env,
			fmt.Sprintf("HELM_REPO_USERNAME=%s", p.Helm.Username),
			fmt.Sprintf("HELM_REPO_PASSWORD=%s", p.Helm.Password),
		)

		cmds = append(cmds, commandPush(p.Chart, p.Helm.RepoUrl))
	}

	if p.Helm.Oci {
		cmds = append(cmds, commandOciSave(p.Chart)) // chart save

		if p.DryRun == false {
			cmds = append(cmds, commandOciPush(p.Chart)) // docker push
		}
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

// helper function to create the helm oci registry login command.
func commandOciLogin(helm Helm) *exec.Cmd {
	return exec.Command(
		helmExe, "registry", "login",
		"-u", helm.Username,
		"-p", helm.Password,
		helm.RegistryUrl,
	)
}

// helper function to create the legacy helm push command.
func commandPush(chart Chart, repoUrl string) *exec.Cmd {
	return exec.Command(
		helmExe, "push",
		"-v", chart.Version,
		chart.Path,
		repoUrl,
	)
}

// helper function to create the helm oci chart save command.
func commandOciSave(chart Chart) *exec.Cmd {
	return exec.Command(
		helmExe, "chart", "save",
		chart.Path,
		chart.OciUrl,
	)
}

// helper function to create the helm oci chart push command.
func commandOciPush(chart Chart) *exec.Cmd {
	return exec.Command(
		helmExe, "chart", "push",
		fmt.Sprintf("%s:%s", chart.OciUrl, chart.Version),
	)
}

func getChartPath(chart Chart) string {
	return path.Join(chart.Path, chart.File)
}

func getChartOciUrl(plugin Plugin) string {
	if plugin.Chart.OciUrl == "" {
		return path.Join(plugin.Helm.RegistryUrl, plugin.Chart.Name)
	}
	return plugin.Chart.OciUrl
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}

func traceBase64(cmd *exec.Cmd) {
	cmdString := strings.Join(cmd.Args, " ")
	fmt.Fprintf(os.Stdout, "+ command: %s\n", base64.StdEncoding.EncodeToString([]byte(cmdString)))
}