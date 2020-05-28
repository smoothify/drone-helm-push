package main

import (
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
			Usage:    "helm registry",
			Required: true,
			EnvVars:  []string{"PLUGIN_HELM_REGISTRY", "PLUGIN_REGISTRY"},
		},
		&cli.BoolFlag{
			Name:    "helm.insecure",
			Usage:   "helm allows insecure registries",
			Value:   false,
			EnvVars: []string{"PLUGIN_INSECURE"},
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
			Name:     "chart.repo",
			Usage:    "chart repo",
			Value:    ".",
			Required: true,
			EnvVars:  []string{"PLUGIN_CHART_REPO", "PLUGIN_REPO"},
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
	plugin := helm_push.Plugin{
		Registry: helm_push.Registry{
			RegistryUrl: c.String("helm.registry"),
			Username:    c.String("helm.username"),
			Password:    c.String("helm.password"),
			Insecure:    c.Bool("helm.insecure"),
		},
		Chart: helm_push.Chart{
			Context: c.String("context"),
			Name:    c.String("chart.name"),
			Path:    c.String("chart.path"),
			File:    c.String("chart.file"),
			Repo:    c.String("chart.repo"),
			Version: cleanVersionString(c.String("chart.version")),
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