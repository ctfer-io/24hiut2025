package main

import (
	"net/url"

	"github.com/ctfer-io/chall-manager/sdk"
	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
	form "github.com/go-playground/form/v4"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	baseFlag = "XSS_Is_3asy_haahahhahahaha"
	port     = 8080
)

type Config struct {
	Hostname string `form:"hostname"`
	Registry string `form:"registry"`
	Image    string `form:"image"`

	IngressAnnotations map[string]string `form:"ingressAnnotations"`
	IngressNamespace   string            `form:"ingressNamespace"`
	IngressLabels      map[string]string `form:"ingressLabels"`
}

func main() {
	sdk.Run(func(req *sdk.Request, resp *sdk.Response, opts ...pulumi.ResourceOption) error {
		conf, err := loadConfig(req.Config.Additional)
		if err != nil {
			return err
		}

		variated := pulumi.Sprintf("24HIUT{%s}", sdk.Variate(req.Config.Identity, baseFlag))

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
					"/app/flag.txt": variated,
				},
			},
			IngressAnnotations: pulumi.ToStringMap(conf.IngressAnnotations),
			IngressNamespace:   pulumi.String(conf.IngressNamespace),
			IngressLabels:      pulumi.ToStringMap(conf.IngressLabels),
		}, opts...)
		if err != nil {
			return err
		}

		resp.ConnectionInfo = pulumi.Sprintf("curl https://%s", cm.URLs.MapIndex(pulumi.String("8080/TCP")))
		resp.Flag = variated.ToStringOutput()
		return nil
	})
}

func loadConfig(additionals map[string]string) (*Config, error) {
	// Default conf
	conf := &Config{
		Hostname: "24hiut25.ctfer.io",
		Image:    "web/blob-cola-1:v0.1.0",
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
	dec := form.NewDecoder()
	if err := dec.Decode(conf, toValues(additionals)); err != nil {
		return nil, err
	}
	return conf, nil
}

func toValues(additionals map[string]string) url.Values {
	vals := make(url.Values, len(additionals))
	for k, v := range additionals {
		vals[k] = []string{v}
	}
	return vals
}
