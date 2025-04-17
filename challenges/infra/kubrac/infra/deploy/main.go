package main

import (
	"fmt"
	"strings"

	"github.com/ctfer-io/chall-manager/sdk"
	"github.com/go-viper/mapstructure/v2"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	netwv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/networking/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	baseFlag = "Kubernetes RBAC are powerfull but dangerous"
	// Could get rick rolled by extracting the value using:
	// kubectl get secret/flag-... --template='{{ index .data "top-secret"}}' | base64 -d
	nggyu = "https://youtu.be/dQw4w9WgXcQ"
	port  = 8080
)

type Config struct {
	Hostname string `mapstructure:"hostname"`
	Registry string `mapstructure:"registry"`
	Image    string `mapstructure:"image"`

	IngressAnnotations map[string]string `mapstructure:"-"`
	IngressNamespace   string            `mapstructure:"-"`
	IngressLabels      map[string]string `mapstructure:"-"`
}

func main() {
	sdk.Run(func(req *sdk.Request, resp *sdk.Response, opts ...pulumi.ResourceOption) error {
		conf, err := loadConfig(req.Ctx, req.Config.Additional)
		if err != nil {
			return err
		}

		// => Namespace
		ns, err := corev1.NewNamespace(req.Ctx, "ns", &corev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: pulumi.StringMap{
					"app.kubernetes.io/component": pulumi.String("kubrac"),
					"app.kubernetes.io/part-of":   pulumi.String("kubrac"),
					// From https://raw.githubusercontent.com/kubernetes/website/main/content/en/examples/security/podsecurity-baseline.yaml
					"pod-security.kubernetes.io/enforce":         pulumi.String("baseline"),
					"pod-security.kubernetes.io/enforce-version": pulumi.String("latest"),
					"pod-security.kubernetes.io/warn":            pulumi.String("baseline"),
					"pod-security.kubernetes.io/warn-version":    pulumi.String("latest"),
				},
			},
		}, opts...)
		if err != nil {
			return err
		}

		labels := pulumi.StringMap{
			"app.kubernetes.io/component": pulumi.String("kubrac"),
			"app.kubernetes.io/part-of":   pulumi.String("kubrac"),
			"app.kubernetes.io/name":      pulumi.String("monitoring"),
		}

		// => Role
		role, err := rbacv1.NewRole(req.Ctx, "role", &rbacv1.RoleArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name().Elem(),
				Labels:    labels,
			},
			Rules: rbacv1.PolicyRuleArray{
				// In the idea of this challenge, there is no role but a kubeconfig,
				// most probably an admin one. The following rules are made to mimic
				// such kubeconfig, but without the permissions to destroy adjacent
				// resources. This isolation is required to ensure no side effects of
				// the RBAC abuse (e.g. cluster role with W permissions).
				rbacv1.PolicyRuleArgs{
					ApiGroups: pulumi.ToStringArray([]string{
						"",
					}),
					Resources: pulumi.ToStringArray([]string{
						"secrets",  // flag
						"pods",     // container list
						"pods/log", // enable
						// following are rabit holes, often looked for, or for
						// completness of basic command `kubectl get all`.
						"configmaps",
						"services",
						"replicationcontrollers",
					}),
					Verbs: pulumi.ToStringArray([]string{
						"get", "list",
					}),
				},
				// For completness of basic command `kubectl get all`.
				rbacv1.PolicyRuleArgs{
					ApiGroups: pulumi.ToStringArray([]string{
						"apps",
					}),
					Resources: pulumi.ToStringArray([]string{
						"daemonsets",
						"deployments",
						"replicasets",
						"statefulsets",
					}),
					Verbs: pulumi.ToStringArray([]string{
						"get", "list",
					}),
				},
				rbacv1.PolicyRuleArgs{
					ApiGroups: pulumi.ToStringArray([]string{
						"autoscaling",
					}),
					Resources: pulumi.ToStringArray([]string{
						"horizontalpodautoscalers",
					}),
					Verbs: pulumi.ToStringArray([]string{
						"get", "list",
					}),
				},
				rbacv1.PolicyRuleArgs{
					ApiGroups: pulumi.ToStringArray([]string{
						"batch",
					}),
					Resources: pulumi.ToStringArray([]string{
						"jobs",
						"cronjobs",
					}),
					Verbs: pulumi.ToStringArray([]string{
						"get", "list",
					}),
				},
			},
		}, opts...)
		if err != nil {
			return err
		}

		// => ServiceAccount
		sa, err := corev1.NewServiceAccount(req.Ctx, "sa", &corev1.ServiceAccountArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name().Elem(),
				Labels:    labels,
			},
		}, opts...)
		if err != nil {
			return err
		}

		// => RoleBinding
		if _, err := rbacv1.NewRoleBinding(req.Ctx, "rb", &rbacv1.RoleBindingArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name().Elem(),
				Labels:    labels,
			},
			RoleRef: rbacv1.RoleRefArgs{
				ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
				Kind:     pulumi.String("Role"),
				Name:     role.Metadata.Name().Elem(),
			},
			Subjects: rbacv1.SubjectArray{
				rbacv1.SubjectArgs{
					Kind:      pulumi.String("ServiceAccount"),
					Name:      sa.Metadata.Name().Elem(),
					Namespace: ns.Metadata.Name().Elem(),
				},
			},
		}, opts...); err != nil {
			return err
		}

		// => Secret
		flag := pulumi.Sprintf("24HIUT{%s}", sdk.Variate(req.Config.Identity, baseFlag, sdk.WithSpecial(true)))
		if _, err = corev1.NewSecret(req.Ctx, "flag", &corev1.SecretArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name().Elem(),
				Labels:    labels,
			},
			StringData: pulumi.StringMap{
				"flag":       flag,
				"top-secret": pulumi.String(nggyu),
			},
			Immutable: pulumi.BoolPtr(true),
		}, opts...); err != nil {
			return err
		}

		// => Deployment
		dep, err := appsv1.NewDeployment(req.Ctx, "monitoring-dep", &appsv1.DeploymentArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name().Elem(),
				Labels:    labels,
			},
			Spec: appsv1.DeploymentSpecArgs{
				Selector: metav1.LabelSelectorArgs{
					MatchLabels: labels,
				},
				Template: corev1.PodTemplateSpecArgs{
					Metadata: metav1.ObjectMetaArgs{
						Namespace: ns.Metadata.Name().Elem(),
						Labels:    labels,
					},
					Spec: corev1.PodSpecArgs{
						ServiceAccountName: sa.Metadata.Name(), // mount ServiceAccount
						ImagePullSecrets: corev1.LocalObjectReferenceArray{
							corev1.LocalObjectReferenceArgs{
								Name: pulumi.String("regcred"),
							},
						},
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name: pulumi.String("monitoring"),
								Image: func(registry, image string) pulumi.StringOutput {
									if registry != "" && !strings.HasSuffix(registry, "/") {
										registry += "/"
									}
									return pulumi.Sprintf("%s%s", registry, image)
								}(conf.Registry, conf.Image),
								Ports: corev1.ContainerPortArray{
									corev1.ContainerPortArgs{
										ContainerPort: pulumi.Int(port),
									},
								},
								Env: corev1.EnvVarArray{
									corev1.EnvVarArgs{
										Name:  pulumi.String("PORT"),
										Value: pulumi.Sprintf("%d", port),
									},
								},
							},
						},
					},
				},
			},
		}, opts...)

		// => Service
		svc, err := corev1.NewService(req.Ctx, "monitoring-svc", &corev1.ServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name().Elem(),
				Labels:    labels,
			},
			Spec: corev1.ServiceSpecArgs{
				ClusterIP: pulumi.String("None"), // Headless, for DNS purposes
				Ports: corev1.ServicePortArray{
					corev1.ServicePortArgs{
						Port: pulumi.Int(port),
					},
				},
				Selector: dep.Spec.ApplyT(func(spec appsv1.DeploymentSpec) map[string]string {
					return spec.Template.Metadata.Labels
				}).(pulumi.StringMapOutput),
			},
		}, opts...)
		if err != nil {
			return err
		}

		// => Ingress
		ing, err := netwv1.NewIngress(req.Ctx, "monitoring-ing", &netwv1.IngressArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name().Elem(),
				Labels:    labels,
				Annotations: pulumi.ToStringMap(map[string]string{
					"pulumi.com/skipAwait": "true",
				}),
			},
			Spec: netwv1.IngressSpecArgs{
				Rules: netwv1.IngressRuleArray{
					netwv1.IngressRuleArgs{
						Host: pulumi.Sprintf("%s.%s", req.Config.Identity, conf.Hostname),
						Http: netwv1.HTTPIngressRuleValueArgs{
							Paths: netwv1.HTTPIngressPathArray{
								netwv1.HTTPIngressPathArgs{
									Path:     pulumi.String("/"),
									PathType: pulumi.String("Prefix"),
									Backend: netwv1.IngressBackendArgs{
										Service: netwv1.IngressServiceBackendArgs{
											Name: svc.Metadata.Name().Elem(),
											Port: netwv1.ServiceBackendPortArgs{
												Number: pulumi.Int(port),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, opts...)
		if err != nil {
			return err
		}

		// => NetworkPolicy to grant all network interactions within the namespace
		if _, err = netwv1.NewNetworkPolicy(req.Ctx, "allow-all-within-namespace", &netwv1.NetworkPolicyArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name().Elem(),
			},
			Spec: netwv1.NetworkPolicySpecArgs{
				PodSelector: metav1.LabelSelectorArgs{}, // Selects all Pods in the namespace
				Ingress: netwv1.NetworkPolicyIngressRuleArray{
					netwv1.NetworkPolicyIngressRuleArgs{
						From: netwv1.NetworkPolicyPeerArray{
							netwv1.NetworkPolicyPeerArgs{
								NamespaceSelector: metav1.LabelSelectorArgs{
									MatchLabels: ns.Metadata.Labels(),
								},
								PodSelector: metav1.LabelSelectorArgs{}, // Allows ingress from any Pod in the same namespace
							},
						},
					},
				},
				Egress: netwv1.NetworkPolicyEgressRuleArray{
					netwv1.NetworkPolicyEgressRuleArgs{
						To: netwv1.NetworkPolicyPeerArray{
							netwv1.NetworkPolicyPeerArgs{
								NamespaceSelector: metav1.LabelSelectorArgs{
									MatchLabels: ns.Metadata.Labels(),
								},
								PodSelector: metav1.LabelSelectorArgs{}, // Allows egress to any Pod in the same namespace
							},
						},
					},
				},
				PolicyTypes: pulumi.ToStringArray([]string{
					"Ingress",
					"Egress",
				}),
			},
		}, opts...); err != nil {
			return err
		}

		// => NetworkPolicy to grant DNS resolution (complex scenarios could require
		// to reach other pods in the namespace, e.g. not a scenario that fits into
		// the sdk.ctfer.io/ExposedMonopod architecture, which then would use headless
		// services so DNS resolution).
		if _, err = netwv1.NewNetworkPolicy(req.Ctx, "allow-kube-dns", &netwv1.NetworkPolicyArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name(),
				Labels: pulumi.StringMap{
					"app.kubernetes.io/component": pulumi.String("chall-manager"),
					"app.kubernetes.io/part-of":   pulumi.String("chall-manager"),
				},
			},
			Spec: netwv1.NetworkPolicySpecArgs{
				PolicyTypes: pulumi.ToStringArray([]string{
					"Egress",
				}),
				PodSelector: metav1.LabelSelectorArgs{},
				Egress: netwv1.NetworkPolicyEgressRuleArray{
					netwv1.NetworkPolicyEgressRuleArgs{
						To: netwv1.NetworkPolicyPeerArray{
							netwv1.NetworkPolicyPeerArgs{
								NamespaceSelector: metav1.LabelSelectorArgs{
									MatchLabels: pulumi.StringMap{
										"kubernetes.io/metadata.name": pulumi.String("kube-system"),
									},
								},
								PodSelector: metav1.LabelSelectorArgs{
									MatchLabels: pulumi.StringMap{
										"k8s-app": pulumi.String("kube-dns"),
									},
								},
							},
						},
						Ports: netwv1.NetworkPolicyPortArray{
							netwv1.NetworkPolicyPortArgs{
								Port:     pulumi.Int(53),
								Protocol: pulumi.String("UDP"),
							},
							netwv1.NetworkPolicyPortArgs{
								Port:     pulumi.Int(53),
								Protocol: pulumi.String("TCP"),
							},
						},
					},
				},
			},
		}, opts...); err != nil {
			return err
		}

		// => NetworkPolicy to deny all scenarios from reaching adjacent namespaces
		if _, err = netwv1.NewNetworkPolicy(req.Ctx, "deny-inter-ns", &netwv1.NetworkPolicyArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name(),
				Labels: pulumi.StringMap{
					"app.kubernetes.io/component": pulumi.String("chall-manager"),
					"app.kubernetes.io/part-of":   pulumi.String("chall-manager"),
				},
			},
			Spec: netwv1.NetworkPolicySpecArgs{
				PodSelector: metav1.LabelSelectorArgs{},
				PolicyTypes: pulumi.ToStringArray([]string{
					"Egress",
				}),
				Egress: netwv1.NetworkPolicyEgressRuleArray{
					netwv1.NetworkPolicyEgressRuleArgs{
						To: netwv1.NetworkPolicyPeerArray{
							netwv1.NetworkPolicyPeerArgs{
								NamespaceSelector: metav1.LabelSelectorArgs{
									MatchExpressions: metav1.LabelSelectorRequirementArray{
										metav1.LabelSelectorRequirementArgs{
											Key:      pulumi.String("kubernetes.io/metadata.name"),
											Operator: pulumi.String("NotIn"),
											Values: pulumi.StringArray{
												ns.Metadata.Name().Elem(),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, opts...); err != nil {
			return err
		}

		// => NetworkPolicy to grant access to Internet IPs (required to download fonts, images, etc.)
		if _, err = netwv1.NewNetworkPolicy(req.Ctx, "allow-internet", &netwv1.NetworkPolicyArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name(),
				Labels: pulumi.StringMap{
					"app.kubernetes.io/component": pulumi.String("chall-manager"),
					"app.kubernetes.io/part-of":   pulumi.String("chall-manager"),
				},
			},
			Spec: netwv1.NetworkPolicySpecArgs{
				PodSelector: metav1.LabelSelectorArgs{},
				PolicyTypes: pulumi.ToStringArray([]string{
					"Egress",
				}),
				Egress: netwv1.NetworkPolicyEgressRuleArray{
					netwv1.NetworkPolicyEgressRuleArgs{
						To: netwv1.NetworkPolicyPeerArray{
							netwv1.NetworkPolicyPeerArgs{
								IpBlock: netwv1.IPBlockArgs{
									Cidr: pulumi.String("0.0.0.0/0"),
									Except: pulumi.ToStringArray([]string{
										"10.0.0.0/8",     // internal Kubernetes cluster IP range
										"172.16.0.0/12",  // common internal IP range
										"192.168.0.0/16", // common internal IP range
									}),
								},
							},
						},
					},
				},
			},
		}, opts...); err != nil {
			return err
		}

		if _, err := netwv1.NewNetworkPolicy(req.Ctx, "allow-from-ingress", &netwv1.NetworkPolicyArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name().Elem(),
				Labels:    svc.Metadata.Labels(),
			},
			Spec: netwv1.NetworkPolicySpecArgs{
				PodSelector: metav1.LabelSelectorArgs{
					MatchLabels: svc.Metadata.Labels(),
				},
				PolicyTypes: pulumi.ToStringArray([]string{
					"Ingress",
				}),
				Ingress: netwv1.NetworkPolicyIngressRuleArray{
					netwv1.NetworkPolicyIngressRuleArgs{
						From: netwv1.NetworkPolicyPeerArray{
							netwv1.NetworkPolicyPeerArgs{
								NamespaceSelector: metav1.LabelSelectorArgs{
									MatchLabels: pulumi.StringMap{
										"kubernetes.io/metadata.name": pulumi.String(conf.IngressNamespace),
									},
								},
								PodSelector: metav1.LabelSelectorArgs{
									MatchLabels: pulumi.ToStringMap(conf.IngressLabels),
								},
							},
						},
						Ports: netwv1.NetworkPolicyPortArray{
							netwv1.NetworkPolicyPortArgs{
								Port: pulumi.Int(port),
							},
						},
					},
				},
			},
		}, opts...); err != nil {
			return err
		}

		// Add a fake PopaCola merch website
		if _, err = appsv1.NewDeployment(req.Ctx, "popacola-merch", &appsv1.DeploymentArgs{
			Metadata: metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name().Elem(),
				Labels: pulumi.StringMap{
					"app.kubernetes.io/component": pulumi.String("kubrac"),
					"app.kubernetes.io/part-of":   pulumi.String("kubrac"),
					"app.kubernetes.io/name":      pulumi.String("popa-cola"),
				},
			},
			Spec: appsv1.DeploymentSpecArgs{
				Selector: metav1.LabelSelectorArgs{
					MatchLabels: pulumi.StringMap{
						"app.kubernetes.io/component": pulumi.String("kubrac"),
						"app.kubernetes.io/part-of":   pulumi.String("kubrac"),
						"app.kubernetes.io/name":      pulumi.String("popa-cola"),
					},
				},
				Template: corev1.PodTemplateSpecArgs{
					Metadata: metav1.ObjectMetaArgs{
						Namespace: ns.Metadata.Name().Elem(),
						Labels: pulumi.StringMap{
							"app.kubernetes.io/component": pulumi.String("kubrac"),
							"app.kubernetes.io/part-of":   pulumi.String("kubrac"),
							"app.kubernetes.io/name":      pulumi.String("popa-cola"),
						},
					},
					Spec: corev1.PodSpecArgs{
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name: pulumi.String("popacola-website"),
								Image: func(registry, image string) pulumi.StringOutput {
									if registry != "" && !strings.HasSuffix(registry, "/") {
										registry += "/"
									}
									return pulumi.Sprintf("%s%s", registry, image)
								}(conf.Registry, "busybox"),
								Command: pulumi.ToStringArray([]string{"/bin/sh", "-c"}),
								Args: pulumi.ToStringArray([]string{`echo "Listening on port 443"
routes='
GET     /products/coke-bottle
GET     /products/t-shirt
GET     /products/cap
POST    /cart/checkout
GET     /popacola
GET     /contact
GET     /products
GET     /about
GET     /faq
POST    /api/order
'
codes="200 200 200 302 404"  # Weighted toward 200
i=0
while true; do
	ip="192.168.1.$(( (RANDOM % 254) + 1 ))"
	route=$(echo "$routes" | sed -n "$(( (RANDOM % 10) + 1 ))p")
	code=$(echo $codes | tr ' ' '\n' | sed -n "$(( (RANDOM % 5) + 1 ))p")
	time=$(awk -v min=0.123 -v max=4.000 'BEGIN{srand(); printf "%.4f", min+rand()*(max-min)}')
	ts=$(date "+%Y/%m/%d - %H:%M:%S")
	echo "[GIN] $ts | $code | ${time}ms | $ip | $route"
	i=$((i+1))
	sleep 10
done`}),
							},
						},
					},
				},
			},
		}, opts...); err != nil {
			return err
		}

		resp.ConnectionInfo = ing.Spec.ApplyT(func(spec netwv1.IngressSpec) string {
			return fmt.Sprintf("curl http://%s", *spec.Rules[0].Host)
		}).(pulumi.StringOutput)
		resp.Flag = flag
		return nil
	})
}

func loadConfig(ctx *pulumi.Context, additionals map[string]string) (*Config, error) {
	// Default conf
	conf := &Config{
		Hostname: "24hiut25.ctfer.io",
		Image:    "infra/kubrac:v0.1.0",
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
