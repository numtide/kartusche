package login

import (
	"fmt"
	"net/url"
	"time"

	"github.com/draganm/kartusche/common/auth"
	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/server"
	"github.com/pkg/browser"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "login",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "with-browser",
			EnvVars: []string{"KARTUSCHE_WITH_BROWSER"},
			Usage:   "when set, CLI will start browser with the auth URL",
		},
	},
	Action: func(c *cli.Context) (err error) {

		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while logging in: %w", err), 1)
			}
		}()

		if c.NArg() != 1 {
			return fmt.Errorf("expected one argument (server url), got %d", c.NArg())
		}

		serverURL := c.Args().First()

		loginStartResponse := &server.LoginStartResponse{}
		err = client.CallAPI(serverURL, "POST", "auth/login", nil, nil, client.JSONDecoder(loginStartResponse), 200)
		if err != nil {
			return fmt.Errorf("while starting login process: %w", err)
		}

		ur, err := url.Parse(loginStartResponse.VerificationURI)
		if err != nil {
			return fmt.Errorf("while parsing verification URI(%s): %w", loginStartResponse.VerificationURI, err)
		}

		q := url.Values{}
		q.Set("request_id", loginStartResponse.TokenRequestID)
		ur.RawQuery = q.Encode()

		fmt.Printf("Please complete authentication flow by visiting %s\n", ur.String())

		if c.Bool("with-browser") {
			go browser.OpenURL(ur.String())
		}

		for {

			var tr server.AccessTokenResponse
			err = client.CallAPI(
				serverURL,
				"POST", "auth/access_token",
				nil,
				client.JSONEncoder(
					server.RequestTokenParameters{
						TokenRequestID: loginStartResponse.TokenRequestID,
					},
				),
				client.JSONDecoder(&tr),
				200,
			)

			if err != nil {
				return fmt.Errorf("while fetching token: %w", err)
			}

			if tr.Error == "authorization_pending" {
				time.Sleep(300 * time.Millisecond)
				continue
			}

			if tr.Error != "" {
				return fmt.Errorf("auth error: %s", tr.Error)
			}

			fmt.Println("auth successful")

			err = auth.StoreTokenForServer(serverURL, tr.AccessToken)
			if err != nil {
				return err
			}

			break

		}
		return nil

	},
}
