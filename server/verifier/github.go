package verifier

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type githubProvider struct {
	clientID     string
	clientSecret string
	organization string
}

func (m *githubProvider) Verify(w http.ResponseWriter, r *http.Request) {

	scheme := r.Header.Get("X-Forwarded-Proto")

	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	verificationURL := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   "/auth/oauth2/callback",
	}

	q := url.Values{}
	q.Set("state", r.URL.Query().Get("request_id"))
	q.Set("client_id", m.clientID)
	q.Set("allow_signup", "false")
	q.Set("redirect_uri", verificationURL.String())
	q.Set("scope", "read:user user:email read:org")

	authorizeURL := &url.URL{
		Scheme:   "https",
		Host:     "github.com",
		Path:     "/login/oauth/authorize",
		RawQuery: q.Encode(),
	}

	http.Redirect(w, r, authorizeURL.String(), 302)
}

type userProfile struct {
	Login                   string      `json:"login"`
	ID                      int         `json:"id"`
	NodeID                  string      `json:"node_id"`
	AvatarURL               string      `json:"avatar_url"`
	GravatarID              string      `json:"gravatar_id"`
	URL                     string      `json:"url"`
	HTMLURL                 string      `json:"html_url"`
	FollowersURL            string      `json:"followers_url"`
	FollowingURL            string      `json:"following_url"`
	GistsURL                string      `json:"gists_url"`
	StarredURL              string      `json:"starred_url"`
	SubscriptionsURL        string      `json:"subscriptions_url"`
	OrganizationsURL        string      `json:"organizations_url"`
	ReposURL                string      `json:"repos_url"`
	EventsURL               string      `json:"events_url"`
	ReceivedEventsURL       string      `json:"received_events_url"`
	Type                    string      `json:"type"`
	SiteAdmin               bool        `json:"site_admin"`
	Name                    string      `json:"name"`
	Company                 string      `json:"company"`
	Blog                    string      `json:"blog"`
	Location                string      `json:"location"`
	Email                   string      `json:"email"`
	Hireable                bool        `json:"hireable"`
	Bio                     string      `json:"bio"`
	TwitterUsername         interface{} `json:"twitter_username"`
	PublicRepos             int         `json:"public_repos"`
	PublicGists             int         `json:"public_gists"`
	Followers               int         `json:"followers"`
	Following               int         `json:"following"`
	CreatedAt               time.Time   `json:"created_at"`
	UpdatedAt               time.Time   `json:"updated_at"`
	PrivateGists            int         `json:"private_gists"`
	TotalPrivateRepos       int         `json:"total_private_repos"`
	OwnedPrivateRepos       int         `json:"owned_private_repos"`
	DiskUsage               int         `json:"disk_usage"`
	Collaborators           int         `json:"collaborators"`
	TwoFactorAuthentication bool        `json:"two_factor_authentication"`
	Plan                    struct {
		Name          string `json:"name"`
		Space         int    `json:"space"`
		Collaborators int    `json:"collaborators"`
		PrivateRepos  int    `json:"private_repos"`
	} `json:"plan"`
}

type UserOrg struct {
	Login            string `json:"login"`
	ID               int    `json:"id"`
	NodeID           string `json:"node_id"`
	URL              string `json:"url"`
	ReposURL         string `json:"repos_url"`
	EventsURL        string `json:"events_url"`
	HooksURL         string `json:"hooks_url"`
	IssuesURL        string `json:"issues_url"`
	MembersURL       string `json:"members_url"`
	PublicMembersURL string `json:"public_members_url"`
	AvatarURL        string `json:"avatar_url"`
	Description      string `json:"description"`
}

func (m *githubProvider) Callback(w http.ResponseWriter, r *http.Request) (ar *AuthResult, err error) {
	w.Write([]byte("authentication successful"))

	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	verificationURL := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   "/auth/oauth2/callback",
	}

	reqParams := r.URL.Query()

	code := reqParams.Get("state")

	defer func() {
		if ar == nil {
			ar = &AuthResult{Code: code}
		}
	}()

	q := url.Values{}
	q.Set("client_id", m.clientID)
	q.Set("client_secret", m.clientSecret)
	q.Set("code", reqParams.Get("code"))
	q.Set("redirect_uri", verificationURL.String())

	tokenURL := &url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "/login/oauth/access_token",
	}

	res, err := http.DefaultClient.Post(tokenURL.String(), "application/x-www-form-urlencoded", strings.NewReader(q.Encode()))
	if err != nil {
		return ar, fmt.Errorf("while getting github token")
	}

	defer res.Body.Close()

	err = res.Request.ParseForm()
	if err != nil {
		return ar, fmt.Errorf("while parsing token form: %w", err)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return ar, fmt.Errorf("while reading token response body: %w", err)
	}

	tq, err := url.ParseQuery(string(b))
	if err != nil {
		return ar, fmt.Errorf("while parsing token response: %w", err)
	}

	authError := tq.Get("error")
	if authError != "" {
		return nil, fmt.Errorf("auth error: %s", authError)
	}

	tkn := tq.Get("access_token")

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("while creating get user request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", tkn))

	res, err = http.DefaultClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("")
	}

	defer res.Body.Close()

	up := &userProfile{}

	err = json.NewDecoder(res.Body).Decode(up)
	if err != nil {
		return nil, fmt.Errorf("while decoding user profile: %w", err)
	}

	req, err = http.NewRequest("GET", "https://api.github.com/user/orgs", nil)
	if err != nil {
		return nil, fmt.Errorf("while creating get user request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", tkn))

	res, err = http.DefaultClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("")
	}

	defer res.Body.Close()

	userOrgs := []UserOrg{}

	err = json.NewDecoder(res.Body).Decode(&userOrgs)
	if err != nil {
		return nil, fmt.Errorf("while decoding user orgs: %w", err)
	}

	for _, uo := range userOrgs {
		if uo.Login == m.organization {
			return &AuthResult{
				Code:  code,
				Email: up.Email,
			}, nil
		}
	}

	return nil, fmt.Errorf("user is not in the requested org")
}

// https://kartusche.netice9.xyz/auth/oauth2/callback
func NewGithubProvider(clientID, clientSecret, organization string) AuthenticationProvider {
	return &githubProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		organization: organization,
	}
}
