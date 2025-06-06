package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ctfer-io/pulumi-ctfd/sdk/v2/go/ctfd"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/pulumiverse/pulumi-fortios/sdk/go/fortios"
	"github.com/pulumiverse/pulumi-fortios/sdk/go/fortios/user"
)

type Input struct {
	Teams []*Team `json:"teams"`
}

type Team struct {
	Name        string    `json:"name"`
	Affiliation string    `json:"affiliation"`
	Players     []*Player `json:"players"`
}

type Player struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func main() {
	b, err := os.ReadFile("players.json")
	if err != nil {
		log.Fatal(err)
	}
	in := Input{}
	if err := json.Unmarshal(b, &in); err != nil {
		log.Fatal(err)
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "players")
		opts := []pulumi.ResourceOption{}
		opts_forti := []pulumi.ResourceOption{}

		users := pulumi.StringMapArray{}.ToStringMapArrayOutput()

		// Create providers
		pv, err := ctfd.NewProvider(ctx, "ctfd-pv", &ctfd.ProviderArgs{
			Url:      pulumi.String(cfg.Require("url")),
			Username: pulumi.String(cfg.Require("username")),
			Password: cfg.RequireSecret("password"),
		})
		if err != nil {
			return err
		}
		opts = append(opts, pulumi.Provider(pv))

		pvforti, err := fortios.NewProvider(ctx, "forti-pv", &fortios.ProviderArgs{
			Hostname: pulumi.String(cfg.Require("forti-address")),
			Token:    cfg.RequireSecret("forti-token"),
			Insecure: pulumi.Bool(true), // trust default cert
		})
		opts_forti = append(opts, pulumi.Provider(pvforti))

		members_grp := user.GroupMemberArray{}

		// Create brackets
		studBkt, err := ctfd.NewBracket(ctx, "students", &ctfd.BracketArgs{
			Name:        pulumi.String("Étudiants"),
			Description: pulumi.String("Bracket des étudiants."),
			Type:        pulumi.String("teams"),
		}, opts...)
		if err != nil {
			return err
		}
		intBkt, err := ctfd.NewBracket(ctx, "interns", &ctfd.BracketArgs{
			Name:        pulumi.String("Internes"),
			Description: pulumi.String("Bracket des internes."),
			Type:        pulumi.String("teams"),
		}, opts...)
		if err != nil {
			return err
		}

		// Create users and teams
		for tid, team := range in.Teams {
			members := pulumi.IDArray{}.ToIDArrayOutput()
			var lastMember *ctfd.User
			bkt := studBkt
			for pid, player := range team.Players {
				// Generate user with random password
				pass, err := random.NewRandomPassword(ctx, fmt.Sprintf("pass-%d-%d", tid, pid), &random.RandomPasswordArgs{
					Length: pulumi.Int(16),
				}, opts...)
				if err != nil {
					return err
				}

				u2, err := user.NewLocal(ctx, fmt.Sprintf("forti-user-%d-%d", tid, pid), &user.LocalArgs{
					Name:    pulumi.String(player.Name),
					EmailTo: pulumi.String(player.Email),
					Passwd:  pass.Result,
					Status:  pulumi.String("enable"),
					Type:    pulumi.String("password"),
				}, opts_forti...)
				if err != nil {
					return errors.Wrapf(err, "team %s, user %s", team.Name, player.Name)
				}

				members_grp = append(members_grp, &user.GroupMemberArgs{
					Name: u2.Name,
				})

				// Companions don't need a CTFd account
				if player.Role == "companion" {
					continue
				}
				// If any member is not a student, go to the interns bracket
				if player.Role != "student" {
					bkt = intBkt
				}

				u, err := ctfd.NewUser(ctx, fmt.Sprintf("team-%d-user-%d", tid, pid), &ctfd.UserArgs{
					Name:     pulumi.String(player.Name),
					Email:    pulumi.String(player.Email),
					Password: pass.Result,
					Type: pulumi.String(func(role string) string {
						switch role {
						case "admin":
							return "admin"
						default:
							return "user"
						}
					}(player.Role)),
				}, opts...)
				if err != nil {
					return errors.Wrapf(err, "team %s, user %s", team.Name, player.Name)
				}

				members = pulumi.All(members, u.ID()).ApplyT(func(all []any) []pulumi.ID {
					return append(all[0].([]pulumi.ID), all[1].(pulumi.ID))
				}).(pulumi.IDArrayOutput)
				lastMember = u

				users = pulumi.All(users, u.Name, pass.Result).ApplyT(func(all []any) []map[string]string {
					users := all[0].([]map[string]string)
					users = append(users, map[string]string{
						"name":     all[1].(string),
						"password": all[2].(string),
					})
					return users
				}).(pulumi.StringMapArrayOutput)
			}

			// Then generate team and assign participants
			pass, err := random.NewRandomPassword(ctx, fmt.Sprintf("pass-%d", tid), &random.RandomPasswordArgs{
				Length: pulumi.Int(16),
			}, opts...)
			if err != nil {
				return err
			}
			if _, err := ctfd.NewTeam(ctx, fmt.Sprintf("team-%d", tid), &ctfd.TeamArgs{
				Name:        pulumi.String(team.Name),
				Email:       lastMember.Email,
				Affiliation: pulumi.String(team.Affiliation),
				Captain: members.ApplyT(func(members []pulumi.ID) pulumi.ID {
					return members[0]
				}).(pulumi.IDOutput),
				Members: members.ApplyT(func(ids []pulumi.ID) []string {
					out := make([]string, 0, len(ids))
					for _, id := range ids {
						out = append(out, string(id))
					}
					return out
				}).(pulumi.StringArrayOutput),
				BracketId: bkt.ID(),
				Password:  pass.Result,
			}, opts...); err != nil {
				return errors.Wrapf(err, "team %s", team.Name)
			}

		}

		// Fortigate Groups
		_, err = user.NewGroup(ctx, "guest-forti", &user.GroupArgs{
			Members:   members_grp,
			Name:      pulumi.String("custom-guest"),
			GroupType: pulumi.String("firewall"),
		}, opts_forti...)
		if err != nil {
			return errors.Wrapf(err, "forti group")
		}

		ctx.Export("players", pulumi.ToSecret(users))

		return nil
	})
}
