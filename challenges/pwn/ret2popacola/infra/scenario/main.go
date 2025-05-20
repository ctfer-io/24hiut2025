package main

import (
	"bytes"
	"net/url"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/ctfer-io/chall-manager/sdk"
	k8s "github.com/ctfer-io/chall-manager/sdk/kubernetes"
	"github.com/go-playground/form/v4"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Config combines all possibile inputs to this recipe.
type Config struct {
	Image    string `form:"image"`
	Hostname string `form:"hostname"`
	Registry string `form:"registry"`

	IngressAnnotations map[string]string `form:"ingressAnnotations"`
	IngressNamespace   string            `form:"ingressNamespace"`
	IngressLabels      map[string]string `form:"ingressLabels"`

	ConnectionInfo string `form:"connectionInfo"`
}

type PortArgs struct {
	Port       int            `form:"port"`
	Protocol   string         `form:"protocol"`
	ExposeType k8s.ExposeType `form:"exposeType"`
}

// Values are used as part of the templating of Config.ConnectionInfo.
type Values struct {
	Ports map[string]string
}

func main() {
	sdk.Run(func(req *sdk.Request, resp *sdk.Response, opts ...pulumi.ResourceOption) error {
		conf, err := loadConfig(req.Config.Additional)
		if err != nil {
			return err
		}

		// Build template ASAP -> fail fast
		citmpl, err := template.New("connectionInfo").
			Funcs(sprig.FuncMap()).
			Parse(conf.ConnectionInfo)
		if err != nil {
			return errors.Wrap(err, "building connection info template")
		}

		// Deploy k8s.ExposedMonopod
		cm, err := k8s.NewExposedMonopod(req.Ctx, "recipe-emp", &k8s.ExposedMonopodArgs{
			Identity: pulumi.String(req.Config.Identity),
			Hostname: pulumi.String(conf.Hostname),
			Label:    pulumi.String(req.Ctx.Stack()),
			Container: k8s.ContainerArgs{
				Image: pulumi.String(func() string {
					if conf.Registry != "" && !strings.HasSuffix(conf.Registry, "/") {
						conf.Registry += "/"
					}
					return conf.Registry + conf.Image
				}()),
				Ports: k8s.PortBindingArray{
					k8s.PortBindingArgs{
						Port:       pulumi.Int(4444),
						ExposeType: k8s.ExposeNodePort,
					},
				},
			},
			IngressAnnotations: pulumi.ToStringMap(conf.IngressAnnotations),
			IngressNamespace:   pulumi.String(conf.IngressNamespace),
			IngressLabels:      pulumi.ToStringMap(conf.IngressLabels),
		}, opts...)
		if err != nil {
			return err
		}

		// Template connection info
		resp.ConnectionInfo = cm.URLs.ApplyT(func(urls map[string]string) (string, error) {
			values := &Values{
				Ports: urls,
			}
			buf := &bytes.Buffer{}
			if err := citmpl.Execute(buf, values); err != nil {
				return "", err
			}
			return buf.String(), nil
		}).(pulumi.StringOutput)

		return nil
	})
}

func loadConfig(additionals map[string]string) (*Config, error) {
	// Default conf
	conf := &Config{
		Hostname: "24hiut25.ctfer.io",
		Image:    "pwn/ret2popacola:v0.1.0",
		ConnectionInfo: `{{- $hostport := index .Ports "4444/TCP" -}}
{{- $parts := splitList ":" $hostport -}}
{{- $host := index $parts 0 -}}
{{- $port := "" -}}
{{- if gt (len $parts) 1 -}}
	{{- $port = index $parts 1 -}}
{{- end -}}
nc {{ $host }} {{ $port }}`,
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
