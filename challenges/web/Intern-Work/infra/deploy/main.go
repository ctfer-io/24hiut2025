package main

import (
	"github.com/ctfer-io/chall-manager/sdk"
	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
	"github.com/mitchellh/mapstructure"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	baseFlag = "M4yb3_1nt3rn_D0nT_H4ck_H4rd3R_tH4n_U"
	port     = 5000
)

type Config struct {
	Hostname string `mapstructure:"hostname"`
	Registry string `mapstructure:"registry"`
	Image    string `mapstructure:"image"`

	IngressAnnotations map[string]string `mapstructure:"ingressAnnotations"`
	IngressNamespace   string            `mapstructure:"ingressNamespace"`
	IngressLabels      map[string]string `mapstructure:"ingressLabels"`
}

func main() {
	sdk.Run(func(req *sdk.Request, resp *sdk.Response, opts ...pulumi.ResourceOption) error {
		conf, err := loadConfig(req.Config.Additional)
		if err != nil {
			return err
		}

		variated := pulumi.Sprintf("24HIUT{%s}", sdk.Variate(req.Config.Identity, baseFlag,
			sdk.WithSpecial(true),
		))

		cm, err := k8s.NewExposedMonopod(req.Ctx, "test", &k8s.ExposedMonopodArgs{
			Identity: pulumi.String(req.Config.Identity),
			Hostname: pulumi.String(conf.Hostname),
			Container: k8s.ContainerArgs{
				Image: pulumi.String(conf.Image),
				Ports: k8s.PortBindingArray{
					k8s.PortBindingArgs{
						Port:       pulumi.Int(port),
						ExposeType: k8s.ExposeIngress,
					},
				},
				Files: pulumi.StringMap{
					"/flag.txt": variated,
				},
			},
			IngressAnnotations: pulumi.ToStringMap(conf.IngressAnnotations),
			IngressNamespace:   pulumi.String(conf.IngressNamespace),
			IngressLabels:      pulumi.ToStringMap(conf.IngressLabels),
		}, opts...)
		if err != nil {
			return err
		}

		resp.ConnectionInfo = pulumi.Sprintf("curl https://%s", cm.URLs.MapIndex(pulumi.String("5000/TCP")))
		resp.Flag = variated.ToStringOutput()
		return nil
	})
}

func loadConfig(additionals map[string]any) (*Config, error) {
	// Default conf
	conf := &Config{
		Hostname: "24hiut25.ctfer.io",
		Image:    "web/intern-work:v0.1.0",
		Registry: "", // keep empty
		// The following fits for a Nginx-based use case, which is the local setup
		IngressAnnotations: map[string]string{
			"kubernetes.io/ingress.class":                  "nginx",
			"nginx.ingress.kubernetes.io/backend-protocol": "HTTP",
			"nginx.ingress.kubernetes.io/ssl-redirect":     "true",
			"nginx.ingress.kubernetes.io/proxy-body-size":  "50m",
		},
		IngressNamespace: "ingress-nginx",
		IngressLabels: map[string]string{
			"app.kubernetes.io/component": "controller",
			"app.kubernetes.io/instance":  "ingress-nginx",
		},
	}

	// Override with additionals
	if err := mapstructure.Decode(additionals, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
