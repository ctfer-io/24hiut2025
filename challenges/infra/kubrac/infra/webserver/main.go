package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:   "Monitoring",
		Usage:  "Monitoring server for Kubernetes local containers",
		Action: run,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				EnvVars: []string{"PORT"},
				Usage:   "port to serves on",
				Value:   8080,
			},
		},
		Authors: []*cli.Author{
			{
				Name:  "Lucas TESSON - PandatiX",
				Email: "lucastesson@protonmail.com",
			},
		},
	}
	gin.SetMode(gin.ReleaseMode)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(ctx *cli.Context) error {
	router := gin.Default()
	router.StaticFile("/", "static/index.html")

	apiv1 := router.Group("/api/v1/")
	apiv1.GET("/containers", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"data": kubectlListContainer(),
		})
	})
	apiv1.GET("/logs", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"data": kubectl_logs(ctx.Query("name")),
		})
	})

	port := ctx.Int("port")
	fmt.Printf("Listening on port %d\n", port)
	return router.Run(fmt.Sprintf(":%d", port))
}

func kubectlListContainer() []string {
	containers, err := kubectl("get pods -o jsonpath='{.items[*].metadata.name}'")
	if err != nil {
		return []string{}
	}
	return strings.Split(containers, " ")
}

func kubectl_logs(name string) string {
	logs, _ := kubectl(fmt.Sprintf("logs %s --tail=20", name))
	return logs
}

// kubectl eval with the following arguments, unsafe to command injection
func kubectl(args string) (string, error) {
	out, err := exec.Command("sh", "-c", fmt.Sprintf("kubectl %s", args)).CombinedOutput()
	return string(out), err
}
