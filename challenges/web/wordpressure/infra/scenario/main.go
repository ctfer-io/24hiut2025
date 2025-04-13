package main

import (
	"fmt"

	"github.com/ctfer-io/chall-manager/sdk"
	"github.com/mitchellh/mapstructure"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	networkingv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/networking/v1"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Additional struct {
	Registry string `json:"registry"`

	ImageWordpress    string `json:"image-wordpress"`
	ImageWordpressCli string `json:"image-wordpress-cli"`
	ImageMySQL        string `json:"image-mysql"`

	Hostname string `json:"hostname"`
	BaseFlag string `json:"base-flag"`
}

var (
	additional Additional
)

const (
	challenge_name     = "vuln-wordpress"
	challenge_category = "web"
)

func main() {
	sdk.Run(func(req *sdk.Request, resp *sdk.Response, opts ...pulumi.ResourceOption) error {

		mysql_labels := pulumi.ToStringMap(map[string]string{
			"identity":  req.Config.Identity,
			"category":  challenge_category,
			"challenge": challenge_name,
			"pod":       "mysql",
		})

		wp_labels := pulumi.ToStringMap(map[string]string{
			"identity":  req.Config.Identity,
			"category":  challenge_category,
			"challenge": challenge_name,
			"pod":       "wordpress",
		})

		//var additional Additional
		err := mapstructure.Decode(req.Config.Additional, &additional)
		if err != nil {
			return err
		}

		err = loadDefaults()
		if err != nil {
			return err
		}

		hostname := fmt.Sprintf("%s.%s", req.Config.Identity, additional.Hostname)
		flag := fmt.Sprintf("24HIUT{%s}", sdk.Variate(req.Config.Identity, additional.BaseFlag))

		// Generate passwords and store them in secrets
		data_root_pass := fmt.Sprintf("%s-mysql-root-pass-%s", challenge_name, req.Config.Identity)
		data_pass := fmt.Sprintf("%s-mysql-pass-%s", challenge_name, req.Config.Identity)
		wp_pass := fmt.Sprintf("%s-wp-pass-%s", challenge_name, req.Config.Identity)

		database_root_pass, err := random.NewRandomPassword(req.Ctx, data_root_pass, &random.RandomPasswordArgs{
			Length:  pulumi.Int(64),
			Special: pulumi.BoolPtr(false),
		}, opts...)
		if err != nil {
			return err
		}

		database_pass, err := random.NewRandomPassword(req.Ctx, data_pass, &random.RandomPasswordArgs{
			Length:  pulumi.Int(64),
			Special: pulumi.BoolPtr(false),
		}, opts...)
		if err != nil {
			return err
		}

		wordpress_pass, err := random.NewRandomPassword(req.Ctx, wp_pass, &random.RandomPasswordArgs{
			Length:  pulumi.Int(32),
			Special: pulumi.BoolPtr(false),
		}, opts...)
		if err != nil {
			return err
		}

		secName := fmt.Sprintf("%s-sec-%s", challenge_name, req.Config.Identity)
		_, err = corev1.NewSecret(req.Ctx, secName, &corev1.SecretArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: wp_labels,
				Name:   pulumi.String(secName),
			},
			Type: pulumi.String("Opaque"),
			StringData: pulumi.ToStringMapOutput(map[string]pulumi.StringOutput{
				"database-root-password": database_pass.Result,
				"database-password":      database_root_pass.Result,
				"wordpress-password":     wordpress_pass.Result,
			}),
		}, opts...)
		if err != nil {
			return err
		}

		// 1. Create MySQL dep and svc
		svcName := fmt.Sprintf("%s-mysql-svc-hl-%s", challenge_name, req.Config.Identity)
		database_svc, err := corev1.NewService(req.Ctx, svcName, &corev1.ServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: mysql_labels,
				Name:   pulumi.String(svcName),
			},
			Spec: corev1.ServiceSpecArgs{
				Type:      pulumi.String("ClusterIP"),
				Selector:  mysql_labels,
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

		depName := fmt.Sprintf("%s-mysql-dep-%s", challenge_name, req.Config.Identity)
		_, err = appsv1.NewDeployment(req.Ctx, depName, &appsv1.DeploymentArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: mysql_labels,
				Name:   pulumi.String(depName),
			},
			Spec: appsv1.DeploymentSpecArgs{
				Selector: metav1.LabelSelectorArgs{
					MatchLabels: mysql_labels,
				},
				Template: corev1.PodTemplateSpecArgs{
					Metadata: metav1.ObjectMetaArgs{
						Labels: mysql_labels,
					},
					Spec: corev1.PodSpecArgs{
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name:  pulumi.String("mysql"),
								Image: pulumi.String(additional.ImageMySQL),
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
												Name: pulumi.String(secName),
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
												Name: pulumi.String(secName),
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

		// 2. Wordpress UI
		svcName = fmt.Sprintf("%s-wp-svc-%s", challenge_name, req.Config.Identity)
		wordpress_svc, err := corev1.NewService(req.Ctx, svcName, &corev1.ServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: wp_labels,
				Name:   pulumi.String(svcName),
			},
			Spec: corev1.ServiceSpecArgs{
				Type:     pulumi.String("ClusterIP"),
				Selector: wp_labels,
				Ports: corev1.ServicePortArray{
					corev1.ServicePortArgs{
						Name:       pulumi.String("wordpress-ui"),
						Port:       pulumi.Int(8000),
						TargetPort: pulumi.Int(80),
					},
				},
			},
		}, opts...)
		if err != nil {
			return err
		}

		// Ingress
		ingName := fmt.Sprintf("%s-wp-ing-%s", challenge_name, req.Config.Identity)
		_, err = networkingv1.NewIngress(req.Ctx, ingName, &networkingv1.IngressArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: wp_labels,
				Name:   pulumi.String(ingName),
				Annotations: pulumi.ToStringMap(map[string]string{
					"pulumi.com/skipAwait": "true",
				}),
			},

			Spec: networkingv1.IngressSpecArgs{
				Rules: networkingv1.IngressRuleArray{
					networkingv1.IngressRuleArgs{
						Host: pulumi.String(hostname),
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
		})
		if err != nil {
			return err
		}

		wp_commands := "cp -rv /plugins/registrationmagic /var/www/html/wp-content/plugins/registrationmagic;"
		wp_commands += "echo configure admin password with ${WORDPRESS_ADMIN_PASSWORD};"
		wp_commands += fmt.Sprintf("wp core install --path='/var/www/html' --url='http://%s' --title='POPACOLA-PREPROD-WEBSITE' --admin_user=admin --admin_password=${WORDPRESS_ADMIN_PASSWORD} --admin_email=admin@popacola.com;", hostname)
		wp_commands += "wp rewrite structure '/%postname%/';"
		wp_commands += "wp plugin activate --path=/var/www/html registrationmagic;"
		wp_commands += fmt.Sprintf("wp post create --post_title='NEW PRODUCT' --post_type=page --post_content='%s' --post_status=draft;", flag)
		wp_commands += "sed -i '1a <!-- admin@popacola.com -->' /var/www/html/wp-content/themes/twentytwentyfive/templates/home.html;"

		depName = fmt.Sprintf("%s-wp-dep-%s", challenge_name, req.Config.Identity)
		_, err = appsv1.NewDeployment(req.Ctx, depName, &appsv1.DeploymentArgs{
			Metadata: metav1.ObjectMetaArgs{
				Labels: wp_labels,
				Name:   pulumi.String(depName),
			},
			Spec: appsv1.DeploymentSpecArgs{
				Selector: metav1.LabelSelectorArgs{
					MatchLabels: wp_labels,
				},
				Template: corev1.PodTemplateSpecArgs{
					Metadata: metav1.ObjectMetaArgs{
						Labels: wp_labels,
					},
					Spec: corev1.PodSpecArgs{
						InitContainers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name:  pulumi.String("wait-db"),
								Image: pulumi.String(additional.ImageMySQL),
								Command: pulumi.ToStringArray([]string{ // TODO
									"/bin/sh",
									"-c",
									"until mysql --host=${DATABASE_URL} --user=wordpress --password=${MYSQL_PASSWORD} --execute=\"SELECT 1;\"; do echo waiting for mysql; sleep 2; done;",
								}),
								Env: corev1.EnvVarArray{
									corev1.EnvVarArgs{
										Name: pulumi.String("MYSQL_PASSWORD"),
										ValueFrom: corev1.EnvVarSourceArgs{
											SecretKeyRef: corev1.SecretKeySelectorArgs{
												Name: pulumi.String(secName),
												Key:  pulumi.String("database-password"),
											},
										},
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("DATABASE_URL"),
										Value: pulumi.Sprintf("%s.%s.svc.cluster.local", database_svc.Metadata.Name().Elem(), database_svc.Metadata.Namespace().Elem()),
									},
								},
							},
							// Initiate file system see https://stackoverflow.com/questions/48623764/using-wordpress-cli-image-on-kubernetes
							corev1.ContainerArgs{
								Name:  pulumi.String("init-wp"),
								Image: pulumi.String(additional.ImageWordpress),
								Args: pulumi.ToStringArray([]string{
									"apache2-foreground",
									"-version",
								}),
								Env: corev1.EnvVarArray{
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_HOST"),
										Value: pulumi.Sprintf("%s.%s.svc.cluster.local", database_svc.Metadata.Name().Elem(), database_svc.Metadata.Namespace().Elem()),
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_USER"),
										Value: pulumi.String("wordpress"),
									},
									corev1.EnvVarArgs{
										Name: pulumi.String("WORDPRESS_DB_PASSWORD"),
										ValueFrom: corev1.EnvVarSourceArgs{
											SecretKeyRef: corev1.SecretKeySelectorArgs{
												Name: pulumi.String(secName),
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
								Name:  pulumi.String("init-wp-cli"),
								Image: pulumi.String(additional.ImageWordpressCli),
								Command: pulumi.ToStringArray([]string{
									"/bin/sh",
									"-c",
									wp_commands,
								}),
								Env: corev1.EnvVarArray{
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_HOST"),
										Value: pulumi.Sprintf("%s.%s.svc.cluster.local", database_svc.Metadata.Name().Elem(), database_svc.Metadata.Namespace().Elem()),
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_USER"),
										Value: pulumi.String("wordpress"),
									},
									corev1.EnvVarArgs{
										Name: pulumi.String("WORDPRESS_DB_PASSWORD"),
										ValueFrom: corev1.EnvVarSourceArgs{
											SecretKeyRef: corev1.SecretKeySelectorArgs{
												Name: pulumi.String(secName),
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
												Name: pulumi.String(secName),
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
								Name:  pulumi.String("wordpress"),
								Image: pulumi.String(additional.ImageWordpress),
								Ports: corev1.ContainerPortArray{
									corev1.ContainerPortArgs{
										ContainerPort: pulumi.Int(80),
									},
								},
								Env: corev1.EnvVarArray{
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_HOST"),
										Value: pulumi.Sprintf("%s.%s.svc.cluster.local", database_svc.Metadata.Name().Elem(), database_svc.Metadata.Namespace().Elem()),
									},
									corev1.EnvVarArgs{
										Name:  pulumi.String("WORDPRESS_DB_USER"),
										Value: pulumi.String("wordpress"),
									},
									corev1.EnvVarArgs{
										Name: pulumi.String("WORDPRESS_DB_PASSWORD"),
										ValueFrom: corev1.EnvVarSourceArgs{
											SecretKeyRef: corev1.SecretKeySelectorArgs{
												Name: pulumi.String(secName),
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
						},
						Volumes: corev1.VolumeArray{
							corev1.VolumeArgs{
								Name: pulumi.String("data"),
								EmptyDir: corev1.EmptyDirVolumeSourceArgs{
									SizeLimit: pulumi.String("2Gi"),
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

		// Export outputs
		resp.ConnectionInfo = pulumi.Sprintf("http://%s", hostname).ToStringOutput()
		resp.Flag = pulumi.String(flag).ToStringOutput()
		return nil
	})
}

func loadDefaults() error {
	if additional.ImageMySQL == "" {
		additional.ImageMySQL = "library/mysql:9.2.0"
	}
	if additional.ImageWordpress == "" {
		additional.ImageWordpress = "library/wordpress:php8.2-apache"
	}
	if additional.ImageWordpressCli == "" {
		additional.ImageWordpressCli = "localhost:5000/ctferio/wordpressure-cli:v0.1.3"
	}

	if additional.Registry != "" {
		additional.ImageMySQL = fmt.Sprintf("%s/%s", additional.Registry, additional.ImageMySQL)
		additional.ImageWordpress = fmt.Sprintf("%s/%s", additional.Registry, additional.ImageWordpress)
		additional.ImageWordpressCli = fmt.Sprintf("%s/%s", additional.Registry, additional.ImageWordpressCli)
	}

	if additional.Hostname == "" {
		additional.Hostname = "localhost"
	}

	if additional.BaseFlag == "" {
		additional.BaseFlag = "Brand_New_Popa_Coola"
	}

	return nil
}
