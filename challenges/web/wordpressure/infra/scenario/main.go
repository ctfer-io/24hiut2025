package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ctfer-io/chall-manager/sdk"
	"github.com/go-playground/form/v4"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	networkingv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/networking/v1"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	baseFlag = "Brand_New_Popa_Coola"
)

type Config struct {
	Hostname          string `form:"hostname"`
	Registry          string `form:"registry"`
	ImageWordpress    string `form:"image-wordpress"`
	ImageWordpressCLI string `form:"image-wordpress-cli"`
	ImageMySQL        string `form:"image-mysql"`

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

		flag := fmt.Sprintf("24HIUT{%s}", sdk.Variate(req.Config.Identity, baseFlag))

		wpLabels := pulumi.StringMap{
			"part": pulumi.String("wordpress"),
			// CTFer.io Chall-Manager labels for filtering
			"chall-manager.ctfer.io/kind":      pulumi.String("custom"),
			"chall-manager.ctfer.io/identity":  pulumi.String(req.Config.Identity),
			"chall-manager.ctfer.io/category":  pulumi.String("web"),
			"chall-manager.ctfer.io/challenge": pulumi.String("wordpressure"),
		}

		mysqlLabels := pulumi.StringMap{
			"part": pulumi.String("mysql"),
			// CTFer.io Chall-Manager labels for filtering
			"chall-manager.ctfer.io/kind":      pulumi.String("custom"),
			"chall-manager.ctfer.io/identity":  pulumi.String(req.Config.Identity),
			"chall-manager.ctfer.io/category":  pulumi.String("web"),
			"chall-manager.ctfer.io/challenge": pulumi.String("wordpressure"),
		}

		// Generate passwords and store them in secrets
		dbRootPass, err := random.NewRandomPassword(req.Ctx, "mysql-root-pass", &random.RandomPasswordArgs{
			Length:  pulumi.Int(64),
			Special: pulumi.BoolPtr(false),
		}, opts...)
		if err != nil {
			return err
		}

		dbPass, err := random.NewRandomPassword(req.Ctx, "mysql-pass", &random.RandomPasswordArgs{
			Length:  pulumi.Int(64),
			Special: pulumi.BoolPtr(false),
		}, opts...)
		if err != nil {
			return err
		}

		wpPass, err := random.NewRandomPassword(req.Ctx, "wp-pass", &random.RandomPasswordArgs{
			Length:  pulumi.Int(32),
			Special: pulumi.BoolPtr(false),
		}, opts...)
		if err != nil {
			return err
		}

		// => Secret, holds all secrets of the scenario
		sec, err := corev1.NewSecret(req.Ctx, "wp-secret", &corev1.SecretArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: wpLabels,
			},
			Type: pulumi.String("Opaque"),
			StringData: pulumi.ToStringMapOutput(map[string]pulumi.StringOutput{
				"database-root-password": dbPass.Result,
				"database-password":      dbRootPass.Result,
				"wordpress-password":     wpPass.Result,
				"000-default.conf": pulumi.String(`<VirtualHost *:8000>
	ServerAdmin webmaster@localhost
	DocumentRoot /var/www/html
	ErrorLog ${APACHE_LOG_DIR}/error.log
	CustomLog ${APACHE_LOG_DIR}/access.log combined
</VirtualHost>
`).ToStringOutput(),
				"ports.conf": pulumi.String("Listen 8000").ToStringOutput(),
			}),
		}, opts...)
		if err != nil {
			return err
		}

		// => BDD (MySQL)
		_, err = appsv1.NewDeployment(req.Ctx, "mysql-dep", &appsv1.DeploymentArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: mysqlLabels,
			},
			Spec: appsv1.DeploymentSpecArgs{
				Selector: metav1.LabelSelectorArgs{
					MatchLabels: mysqlLabels,
				},
				Template: corev1.PodTemplateSpecArgs{
					Metadata: metav1.ObjectMetaArgs{
						Labels: mysqlLabels,
					},
					Spec: corev1.PodSpecArgs{
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name: pulumi.String("mysql"),
								Image: pulumi.String(func() string {
									if conf.Registry != "" && !strings.HasSuffix(conf.Registry, "/") {
										conf.Registry += "/"
									}
									return conf.Registry + conf.ImageMySQL
								}()),
								Ports: corev1.ContainerPortArray{
									corev1.ContainerPortArgs{
										ContainerPort: pulumi.Int(3306),
									},
								},
								Env: corev1.EnvVarArray{
									corev1.EnvVarArgs{
										Name: pulumi.String("MYSQL_ROOT_PASSWORD"),
										ValueFrom: corev1.EnvVarSourceArgs{
											SecretKeyRef: corev1.SecretKeySelectorArgs{
												Name: sec.Metadata.Name(),
												Key:  pulumi.String("database-root-password"),
											},
										},
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("MYSQL_DATABASE"),
										Value: pulumi.String("wordpress"),
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("MYSQL_USER"),
										Value: pulumi.String("wordpress"),
									},
									corev1.EnvVarArgs{
										Name: pulumi.String("MYSQL_PASSWORD"),
										ValueFrom: corev1.EnvVarSourceArgs{
											SecretKeyRef: corev1.SecretKeySelectorArgs{
												Name: sec.Metadata.Name(),
												Key:  pulumi.String("database-password"),
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

		mysqlSvc, err := corev1.NewService(req.Ctx, "mysql-svc-hl", &corev1.ServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: mysqlLabels,
			},
			Spec: corev1.ServiceSpecArgs{
				Type:      pulumi.String("ClusterIP"),
				Selector:  mysqlLabels,
				ClusterIP: pulumi.String("None"), // headless
				Ports: corev1.ServicePortArray{
					corev1.ServicePortArgs{
						Name:       pulumi.String("mysql"),
						Port:       pulumi.Int(3306),
						TargetPort: pulumi.Int(3306),
					},
				},
			},
		}, opts...)
		if err != nil {
			return err
		}

		// WordPress commands
		wp_commands := "cp -rv /plugins/registrationmagic /var/www/html/wp-content/plugins/registrationmagic;"
		wp_commands += "echo configure admin password with ${WORDPRESS_ADMIN_PASSWORD};"
		wp_commands += fmt.Sprintf("wp core install --path='/var/www/html' --url='http://%s' --title='POPACOLA-PREPROD-WEBSITE' --admin_user=admin --admin_password=${WORDPRESS_ADMIN_PASSWORD} --admin_email=admin@popacola.com;", conf.Hostname)
		wp_commands += "wp rewrite structure '/%postname%/';"
		wp_commands += "wp plugin activate --path=/var/www/html registrationmagic;"
		wp_commands += fmt.Sprintf("wp post create --post_title='NEW PRODUCT' --post_type=page --post_content='%s' --post_status=draft;", flag)
		wp_commands += "sed -i '1a <!-- admin@popacola.com -->' /var/www/html/wp-content/themes/twentytwentyfive/templates/home.html;"

		// => Deployment
		_, err = appsv1.NewDeployment(req.Ctx, "wp-dep", &appsv1.DeploymentArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: wpLabels,
			},
			Spec: appsv1.DeploymentSpecArgs{
				Selector: metav1.LabelSelectorArgs{
					MatchLabels: wpLabels,
				},
				Template: corev1.PodTemplateSpecArgs{
					Metadata: metav1.ObjectMetaArgs{
						Labels: wpLabels,
					},
					Spec: corev1.PodSpecArgs{
						InitContainers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name: pulumi.String("wait-db"),
								Image: pulumi.String(func() string {
									if conf.Registry != "" && !strings.HasSuffix(conf.Registry, "/") {
										conf.Registry += "/"
									}
									return conf.Registry + conf.ImageMySQL
								}()),
								Command: pulumi.ToStringArray([]string{
									"/bin/sh",
									"-c",
									"until mysql --host=${DATABASE_URL} --user=wordpress --password=${MYSQL_PASSWORD} --execute=\"SELECT 1;\"; do echo waiting for mysql; sleep 2; done;",
								}),
								Env: corev1.EnvVarArray{
									corev1.EnvVarArgs{
										Name: pulumi.String("MYSQL_PASSWORD"),
										ValueFrom: corev1.EnvVarSourceArgs{
											SecretKeyRef: corev1.SecretKeySelectorArgs{
												Name: sec.Metadata.Name(),
												Key:  pulumi.String("database-password"),
											},
										},
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("DATABASE_URL"),
										Value: mysqlSvc.Metadata.Name(),
									},
								},
							},
							// Initiate file system see https://stackoverflow.com/questions/48623764/using-wordpress-cli-image-on-kubernetes
							corev1.ContainerArgs{
								Name: pulumi.String("init-wp"),
								Image: pulumi.String(func() string {
									if conf.Registry != "" && !strings.HasSuffix(conf.Registry, "/") {
										conf.Registry += "/"
									}
									return conf.Registry + conf.ImageWordpress
								}()),
								Args: pulumi.ToStringArray([]string{
									"apache2-foreground",
									"-version",
								}),
								Env: corev1.EnvVarArray{
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_HOST"),
										Value: mysqlSvc.Metadata.Name(),
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_USER"),
										Value: pulumi.String("wordpress"),
									},
									corev1.EnvVarArgs{
										Name: pulumi.String("WORDPRESS_DB_PASSWORD"),
										ValueFrom: corev1.EnvVarSourceArgs{
											SecretKeyRef: corev1.SecretKeySelectorArgs{
												Name: sec.Metadata.Name(),
												Key:  pulumi.String("database-password"),
											},
										},
									},
									corev1.EnvVarArgs{

										Name:  pulumi.String("WORDPRESS_DB_NAME"),
										Value: pulumi.String("wordpress"),
									},
								},
								VolumeMounts: corev1.VolumeMountArray{
									corev1.VolumeMountArgs{
										Name:      pulumi.String("data"),
										MountPath: pulumi.String("/var/www/html"),
									},
								},
								SecurityContext: corev1.SecurityContextArgs{
									RunAsUser: pulumi.Int(33),
								},
							},
							corev1.ContainerArgs{
								Name: pulumi.String("init-wp-cli"),
								Image: pulumi.String(func() string {
									if conf.Registry != "" && !strings.HasSuffix(conf.Registry, "/") {
										conf.Registry += "/"
									}
									return conf.Registry + conf.ImageWordpressCLI
								}()),
								Command: pulumi.ToStringArray([]string{
									"/bin/sh",
									"-c",
									wp_commands,
								}),
								Env: corev1.EnvVarArray{
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_HOST"),
										Value: mysqlSvc.Metadata.Name(),
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_USER"),
										Value: pulumi.String("wordpress"),
									},
									corev1.EnvVarArgs{
										Name: pulumi.String("WORDPRESS_DB_PASSWORD"),
										ValueFrom: corev1.EnvVarSourceArgs{
											SecretKeyRef: corev1.SecretKeySelectorArgs{
												Name: sec.Metadata.Name(),
												Key:  pulumi.String("database-password"),
											},
										},
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_NAME"),
										Value: pulumi.String("wordpress"),
									},
									corev1.EnvVarArgs{
										Name: pulumi.String("WORDPRESS_ADMIN_PASSWORD"),
										ValueFrom: corev1.EnvVarSourceArgs{
											SecretKeyRef: corev1.SecretKeySelectorArgs{
												Name: sec.Metadata.Name(),
												Key:  pulumi.String("wordpress-password"),
											},
										},
									},
								},
								VolumeMounts: corev1.VolumeMountArray{
									corev1.VolumeMountArgs{
										Name:      pulumi.String("data"),
										MountPath: pulumi.String("/var/www/html"),
									},
								},
								SecurityContext: corev1.SecurityContextArgs{
									RunAsUser: pulumi.Int(33),
								},
							},
						},
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name: pulumi.String("wordpress"),
								Image: pulumi.String(func() string {
									if conf.Registry != "" && !strings.HasSuffix(conf.Registry, "/") {
										conf.Registry += "/"
									}
									return conf.Registry + conf.ImageWordpress
								}()),
								Ports: corev1.ContainerPortArray{
									corev1.ContainerPortArgs{
										ContainerPort: pulumi.Int(8000),
									},
								},
								Env: corev1.EnvVarArray{
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_HOST"),
										Value: mysqlSvc.Metadata.Name(),
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_USER"),
										Value: pulumi.String("wordpress"),
									},
									corev1.EnvVarArgs{
										Name: pulumi.String("WORDPRESS_DB_PASSWORD"),
										ValueFrom: corev1.EnvVarSourceArgs{
											SecretKeyRef: corev1.SecretKeySelectorArgs{
												Name: sec.Metadata.Name(),
												Key:  pulumi.String("database-password"),
											},
										},
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_NAME"),
										Value: pulumi.String("wordpress"),
									},
								},
								VolumeMounts: corev1.VolumeMountArray{
									corev1.VolumeMountArgs{
										Name:      pulumi.String("data"),
										MountPath: pulumi.String("/var/www/html"),
									},
									corev1.VolumeMountArgs{
										Name:      pulumi.String("config"),
										MountPath: pulumi.String("/etc/apache2/sites-enabled/000-default.conf"),
										SubPath:   pulumi.String("000-default.conf"), // force listen to 8000
										ReadOnly:  pulumi.Bool(true),
									},
									corev1.VolumeMountArgs{
										Name:      pulumi.String("config"),
										MountPath: pulumi.String("/etc/apache2/ports.conf"),
										SubPath:   pulumi.String("ports.conf"), // force listen to 8000
										ReadOnly:  pulumi.Bool(true),
									},
								},
								SecurityContext: corev1.SecurityContextArgs{
									RunAsUser: pulumi.Int(33),
								},
							},
						},
						Volumes: corev1.VolumeArray{
							corev1.VolumeArgs{
								Name: pulumi.String("data"),
								EmptyDir: corev1.EmptyDirVolumeSourceArgs{
									SizeLimit: pulumi.String("2Gi"),
								},
							},
							corev1.VolumeArgs{
								Name: pulumi.String("config"),
								Secret: corev1.SecretVolumeSourceArgs{
									SecretName: sec.Metadata.Name(),
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

		// => Service (expose dep)
		wordpress_svc, err := corev1.NewService(req.Ctx, "wp-svc", &corev1.ServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: wpLabels,
			},
			Spec: corev1.ServiceSpecArgs{
				Type:     pulumi.String("ClusterIP"),
				Selector: wpLabels,
				Ports: corev1.ServicePortArray{
					corev1.ServicePortArgs{
						Name:       pulumi.String("wordpress-ui"),
						Port:       pulumi.Int(8000),
						TargetPort: pulumi.Int(8000),
					},
				},
			},
		}, opts...)
		if err != nil {
			return err
		}

		// => Ingress (expose for outer networking)
		ing, err := networkingv1.NewIngress(req.Ctx, "wp-ing", &networkingv1.IngressArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: wpLabels,
				Annotations: pulumi.ToStringMap(map[string]string{
					"pulumi.com/skipAwait": "true",
				}),
			},
			Spec: networkingv1.IngressSpecArgs{
				Rules: networkingv1.IngressRuleArray{
					networkingv1.IngressRuleArgs{
						Host: pulumi.Sprintf("%s.%s", req.Config.Identity, conf.Hostname),
						Http: networkingv1.HTTPIngressRuleValueArgs{
							Paths: networkingv1.HTTPIngressPathArray{
								networkingv1.HTTPIngressPathArgs{
									Path:     pulumi.String("/"),
									PathType: pulumi.String("Prefix"),
									Backend: networkingv1.IngressBackendArgs{
										Service: networkingv1.IngressServiceBackendArgs{
											Name: wordpress_svc.Metadata.Name().Elem(),
											Port: networkingv1.ServiceBackendPortArgs{
												Name: pulumi.String("wordpress-ui"),
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

		if _, err := networkingv1.NewNetworkPolicy(req.Ctx, "ntp-internal-wordpress-mysql", &networkingv1.NetworkPolicyArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: mysqlLabels,
			},
			Spec: networkingv1.NetworkPolicySpecArgs{
				PodSelector: metav1.LabelSelectorArgs{
					MatchLabels: mysqlLabels,
				},
				PolicyTypes: pulumi.ToStringArray([]string{
					"Ingress",
				}),
				Ingress: networkingv1.NetworkPolicyIngressRuleArray{
					networkingv1.NetworkPolicyIngressRuleArgs{
						From: networkingv1.NetworkPolicyPeerArray{
							networkingv1.NetworkPolicyPeerArgs{
								PodSelector: metav1.LabelSelectorArgs{
									MatchLabels: wpLabels,
								},
							},
						},
						Ports: networkingv1.NetworkPolicyPortArray{
							networkingv1.NetworkPolicyPortArgs{
								Port: pulumi.Int(3306),
							},
						},
					},
				},
			},
		}, opts...); err != nil {
			return err
		}

		if _, err := networkingv1.NewNetworkPolicy(req.Ctx, "ntp-external-ingress-wordpress", &networkingv1.NetworkPolicyArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: wpLabels,
			},
			Spec: networkingv1.NetworkPolicySpecArgs{
				PodSelector: metav1.LabelSelectorArgs{
					MatchLabels: wpLabels,
				},
				PolicyTypes: pulumi.ToStringArray([]string{
					"Ingress",
				}),
				Ingress: networkingv1.NetworkPolicyIngressRuleArray{
					networkingv1.NetworkPolicyIngressRuleArgs{
						From: networkingv1.NetworkPolicyPeerArray{
							networkingv1.NetworkPolicyPeerArgs{
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
						Ports: networkingv1.NetworkPolicyPortArray{
							networkingv1.NetworkPolicyPortArgs{
								Port: pulumi.Int(8000),
							},
						},
					},
				},
			},
		}, opts...); err != nil {
			return err
		}
		// Export outputs
		resp.ConnectionInfo = pulumi.Sprintf("https://%s", ing.Spec.Rules().Index(pulumi.Int(0)).Host()).ToStringOutput()
		resp.Flag = pulumi.String(flag).ToStringOutput()
		return nil
	})
}

func loadConfig(additionals map[string]string) (*Config, error) {
	// Default conf
	conf := &Config{
		Hostname:          "24hiut2025.ctfer.io",
		ImageWordpress:    "library/wordpress:php8.2-apache",
		ImageWordpressCLI: "web/wordpressure-cli:v0.1.0",
		ImageMySQL:        "library/mysql:9.2.0",
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
