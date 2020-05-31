package main

import (
	"errors"
	"github.com/smoothify/drone-helm-push/pkg/helm_push"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"path"
	"regexp"
)

var (
	version = "unknown"
)

func main() {
	// Load env-file if it exists first
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		godotenv.Load(env)
	}

	app := cli.NewApp()
	app.Name = "helm-push plugin"
	app.Usage = "helm-push plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "dry-run",
			Usage:   "dry run disables helm push",
			EnvVars: []string{"PLUGIN_DRY_RUN"},
		},
		&cli.StringFlag{
			Name:    "helm.username",
			Usage:   "helm username",
			EnvVars: []string{"PLUGIN_HELM_USERNAME", "PLUGIN_USERNAME"},
		},
		&cli.StringFlag{
			Name:    "helm.password",
			Usage:   "helm password",
			EnvVars: []string{"PLUGIN_HELM_PASSWORD", "PLUGIN_PASSWORD"},
		},
		&cli.StringFlag{
			Name:     "helm.registry",
			Usage:    "helm oci registry",
			EnvVars:  []string{"PLUGIN_HELM_REGISTRY", "PLUGIN_REGISTRY"},
		},
		&cli.StringFlag{
			Name:     "helm.repo",
			Usage:    "helm legacy repo",
			EnvVars:  []string{"PLUGIN_HELM_REPO", "PLUGIN_REPO"},
		},
		&cli.BoolFlag{
			Name:    "helm.insecure",
			Usage:   "helm allows insecure registries",
			Value:   false,
			EnvVars: []string{"PLUGIN_INSECURE"},
		},
		&cli.BoolFlag{
			Name:    "helm.oci",
			Usage:   "helm enable oci",
			Value:   false,
			EnvVars: []string{"PLUGIN_HELM_OCI", "PLUGIN_OCI"},
		},
		&cli.BoolFlag{
			Name:    "helm.legacy",
			Usage:   "helm enable legacy",
			Value:   true,
			EnvVars: []string{"PLUGIN_HELM_LEGACY", "PLUGIN_LEGACY"},
		},
		&cli.StringFlag{
			Name:    "context",
			Usage:   "context",
			Value:   ".",
			EnvVars: []string{"PLUGIN_CONTEXT"},
		},
		&cli.StringFlag{
			Name:     "chart.name",
			Usage:    "chart name",
			Required: true,
			EnvVars:  []string{"PLUGIN_CHART_NAME"},
		},
		&cli.StringFlag{
			Name:    "chart.path",
			Usage:   "chart path",
			Value:   ".",
			EnvVars: []string{"PLUGIN_CHART_PATH"},
		},
		&cli.StringFlag{
			Name:    "chart.file",
			Usage:   "chart file",
			Value:   "Chart.yaml",
			EnvVars: []string{"PLUGIN_CHART_FILE"},
		},
		&cli.StringFlag{
			Name:    "chart.url",
			Usage:   "chart url",
			EnvVars: []string{"PLUGIN_CHART_URL"},
		},
		&cli.StringFlag{
			Name:    "chart.oci-url",
			Usage:   "chart oci url",
			EnvVars: []string{"PLUGIN_CHART_OCI_URL"},
		},
		&cli.StringFlag{
			Name:    "chart.version",
			Usage:   "chart version",
			EnvVars: []string{"PLUGIN_CHART_VERSION"},
			FilePath: ".release-version",
		},
		&cli.BoolFlag{
			Name:    "error-no-release",
			Usage:   "error on no new release",
			EnvVars: []string{"PLUGIN_ERROR_NO_RELEASE"},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func getExecDir() string {
	ex, err := os.Executable()
	if err != nil {
		logrus.Fatal(err)
	}
	dir := path.Dir(ex)
	return dir
}

func run(c *cli.Context) error {
	if c.Bool("helm.oci") && c.String("helm.registry") == "" {
		return errors.New("when helm.oci is enabled, a helm registry must be supplied")
	}
	if c.Bool("helm.legacy") && c.String("helm.repo") == "" {
		return errors.New("when helm.legacy is enabled, a helm repo must be supplied")
	}

	plugin := helm_push.Plugin{
		Helm: helm_push.Helm{
			RegistryUrl: c.String("helm.registry"),
			RepoUrl:     c.String("helm.repo"),
			Username:    c.String("helm.username"),
			Password:    c.String("helm.password"),
			Insecure:    c.Bool("helm.insecure"),
			Oci:         c.Bool("helm.oci"),
			Legacy:      c.Bool("helm.legacy"),
		},
		Chart: helm_push.Chart{
			Context: c.String("context"),
			Name:    c.String("chart.name"),
			Path:    c.String("chart.path"),
			File:    c.String("chart.file"),
			Version: cleanVersionString(c.String("chart.version")),
			OciUrl:  c.String("chart.oci-url"),
		},
		DryRun:         c.Bool("dry-run"),
		ErrorNoRelease: c.Bool("error-no-release"),
	}

	return plugin.Exec()
}

func cleanVersionString(s string) string {
	re := regexp.MustCompile(`[\r\n ]`)
	return re.ReplaceAllString(s, "")
}