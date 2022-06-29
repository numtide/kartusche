package login

import (
	"fmt"
	"net/url"
	"time"

	"github.com/draganm/kartusche/command/server"
	"github.com/draganm/kartusche/common/auth"
	"github.com/draganm/kartusche/common/client"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "login",
	Flags: []cli.Flag{},
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
		err = client.CallAPI(serverURL, "POST", "auth/login", nil, loginStartResponse, 200)
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

		for {

			var tr server.AccessTokenResponse
			err = client.CallAPI(
				serverURL,
				"POST", "auth/access_token",
				client.JSONEncoder(
					server.RequestTokenParameters{
						TokenRequestID: loginStartResponse.TokenRequestID,
					},
				),
				&tr,
				200,
			)

			if err != nil {
				return fmt.Errorf("while fetching token: %w", err)
			}

			if tr.Error == "authorization_pending" {
				time.Sleep(300 * time.Millisecond)
				continue
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
