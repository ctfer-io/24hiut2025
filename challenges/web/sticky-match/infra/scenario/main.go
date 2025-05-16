package main

import (
	"github.com/ctfer-io/chall-manager/sdk"
	"github.com/ctfer-io/chall-manager/sdk/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	sdk.Run(func(req *sdk.Request, resp *sdk.Response, opts ...pulumi.ResourceOption) error {
		cm, err := kubernetes.NewExposedMonopod(req.Ctx, "sticky-match", &kubernetes.ExposedMonopodArgs{
			Container: kubernetes.ContainerArgs{
				Image: pulumi.String("web/sticky-match:v0.1.0"),
				Ports: kubernetes.PortBindingArray{
					kubernetes.PortBindingArgs{
						Port:       pulumi.Int(80),
						ExposeType: kubernetes.ExposeIngress,
					},
				},
			},
			Hostname: pulumi.String("24hiut2025.ctfer.io"),
			Identity: pulumi.String(req.Config.Identity),
		}, opts...)
		if err != nil {
			return err
		}

		resp.ConnectionInfo = pulumi.Sprintf("curl -v https://%s", cm.URLs.MapIndex(pulumi.String("80/TCP")))
		return nil
	})
}
