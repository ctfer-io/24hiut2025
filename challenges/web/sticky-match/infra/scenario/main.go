package main

import (
	"github.com/ctfer-io/chall-manager/sdk"
	"github.com/ctfer-io/chall-manager/sdk/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	sdk.Run(func(req *sdk.Request, resp *sdk.Response, opts ...pulumi.ResourceOption) error {
		cm, err := kubernetes.NewExposedMonopod(req.Ctx, &kubernetes.ExposedMonopodArgs{
			Image:      pulumi.String("localhost:5000/ctferio/sticky-match:latest"), // challenge Docker image
			Port:       pulumi.Int(80),                          					 // pod listens on port 8080
			ExposeType: kubernetes.ExposeIngress,                  					 // expose the challenge through an ingress (HTTP)
			Hostname:   pulumi.String("24hiut2025.ctfer.io"),        				 // CTF hostname
			Identity:   pulumi.String(req.Config.Identity),        					 // identity will be prepended to hostname
		}, opts...)
		if err != nil {
			return err
		}

		resp.ConnectionInfo = pulumi.Sprintf("curl -v https://%s", cm.URL) // a simple web server
		return nil
	})
}
