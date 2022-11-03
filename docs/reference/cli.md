# Kartusche CLI


## Development
### `develop`

Starts a development server that reloads code from the current project and runs tests automatically.
HTTP clients can connect to <http://localhost:5001>.

#### Options

* `--addr` (env `$KARTUSCHE_ADDR`): `[<hostname|ip>]:<port>` where the developer server will bind the HTTP

### `init`

Creates a new kartusche project in the current folder.

### `test`

Runs all kartusche tests in the current project.


## Hosting
### `server`

Starts a (production) kartusche server.

Listens to two ports, one for the kartusche API and one for serving HTTP requests for the deployed kartusches.

#### Options
* `--controller-addr` (env `$CONTROLLER_ADDR`):              (default: ":3003"): Address where kartusche server will serve API.
* `--kartusches-addr` value              (default: ":3002") [$KARTUSCHES_ADDR]: Address where kartusche server will serve HTTP requests to kartusches.
* `--work-dir` value                     (default: "work") [$WORK_DIR]: Directory where the state of the server and kartusches will be stored.
* `--auth-provider` value                (default: "mock") [$AUTH_PROVIDER]: Name of the Authentication provider for API. Possible values are `mock` and `github`.
* `--oauth2-github-client-id` value       [$OAUTH2_GITHUB_CLIENT_ID]: OAuth2 client id for SSO.
* `--oauth2-github-client-secret` value   [$OAUTH2_GITHUB_CLIENT_SECRET] OAuth2 client secret for SSO.
* `--oauth2-github-organization` value    [$OAUTH2_GITHUB_ORGANIZATION] Members of this Github org will be allowed to use kartusche API.
* `--kartusche-domain value`             (default: "127.0.0.1.nip.io") [$KARTUSCHE_DOMAIN]: Top level DNS domain for serving kartusches. E.g. kartusche with the name `test` will be served under <https://test.your.domain>.


## auth
## clone
## info
## ls
## remote
## rm
## update
## upload
