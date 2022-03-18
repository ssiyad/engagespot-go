package engagespot

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
)

const ENDPOINT = "https://api.engagespot.co/v3/"
const DEVICE_TYPE = "ios"

type config struct {
	enableHmac bool
}

// https://documentation.engagespot.co/docs/rest-api#tag/Notifications/paths/~1v3~1notifications/post
// construct a notification schema
// title is required
type schema struct {
	Title   string `json:"title"`
	Message string `json:"message,omitempty"`
	Url     string `json:"url,omitempty"`
	Icon    string `json:"icon,omitempty"`
}

// override have the following fields
// channels
// Array of strings
// Specify the channels through which this notification should be delivered. See Channels to get the
// complete list of supported channels.
// sendgrid_email
// object
// Overrides Sendgrid configuration specified in your Engagespot dashboard. This is considered only
// if you've enabled Sendgrid email provider. The first property _config is used by Engagespot to
// override your Sendgrid default configuration specified in the dashboard. Along with _config
// (Not inside _config), You can pass any property as supported by the Sendgrid's mail send API.
// smtp_email
// object
// Overrides SMTP Provider configurations specified in your Engagespot dashboard. This is considered
// only if you have enabled SMTP Email Provider.
type override struct {
	Channels []string `json:"channels,omitempty"`
}

// AddChannel is a method to override notification channels and resets any set configuration
// on first insertion
func (o *override) AddChannel(channel string) {
	o.Channels = append(o.Channels, channel)
}

// https://documentation.engagespot.co/docs/rest-api#tag/Notifications/paths/~1v3~1notifications/post
// represents a notification schema as defined above
// notification and recipients are required
type notification struct {
	*client
	Notification *schema   `json:"notification"`
	Recipients   []string  `json:"recipients"`
	Category     string    `json:"category,omitempty"`
	Override     *override `json:"override,omitempty"`
}

// SetMessage can be used to set notification message
func (n *notification) SetMessage(message string) (*notification, error) {
	if message == "" {
		return nil, errors.New("empty message string")
	}
	n.Notification.Message = message
	return n, nil
}

// SetUrl can be used to set callback url
func (n *notification) SetUrl(url string) (*notification, error) {
	if url == "" {
		return nil, errors.New("empty url string")
	}
	n.Notification.Url = url
	return n, nil
}

// SetIcon can be used to set notification icon
func (n *notification) SetIcon(iconUrl string) (*notification, error) {
	if iconUrl == "" {
		return nil, errors.New("empty icon url string")
	}
	n.Notification.Icon = iconUrl
	return n, nil
}

// SetCategory can be used to set notification category. If category doesn't exist, it will be created
func (n *notification) SetCategory(category string) (*notification, error) {
	if category == "" {
		return nil, errors.New("empty category string")
	}
	n.Category = category
	return n, nil
}

// AddRecipient can be used to add a recipient to the list. If none is present during send, an error will be thrown
func (n *notification) AddRecipient(recipient string) (*notification, error) {
	if recipient == "" {
		return nil, errors.New("empty recipient string")
	}
	n.Recipients = append(n.Recipients, recipient)
	return n, nil
}

// used to check if enough recipients are present
func (n *notification) hasEnoughRecipients() bool {
	return len(n.Recipients) > 0
}

// send a notification
func (n *notification) Send() (*http.Response, error) {
	if !n.hasEnoughRecipients() {
		return nil, errors.New("not enough recipients")
	}
	return n.client.Send(n)
}

// base struct of client. contain an http client used to communicate with the API
type client struct {
	apiKey     string
	apiSecret  string
	config     config
	httpClient *http.Client
}

// NewEngagespotClient can be used to create a client which can then be used to create
// and send notifications
func NewEngagespotClient(apiKey, apiSecret string) *client {
	httpClient := &http.Client{}

	client := &client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		config:     config{},
		httpClient: httpClient,
	}

	return client
}

// NewNotification can be used to create a notification item which can later be sent by using .Send()
func (c *client) NewNotification(title string) (*notification, error) {
	n := &schema{
		Title: title,
	}
	o := &override{}

	notification := &notification{
		Notification: n,
		Override:     o,
		client:       c,
	}

	return notification, nil
}

// EnableHmac can be used to enable an extra layer of security.
// Read more: https://documentation.engagespot.co/docs/HMAC-authentication/enabling-HMAC-authentication
func (c *client) EnableHmac() *client {
	c.config.enableHmac = true
	return c
}

// basic method to call the API using already defined http client. credentials are set here
func (c *client) call(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")

	req.Header.Add("X-ENGAGESPOT-API-KEY", c.apiKey)
	req.Header.Add("X-ENGAGESPOT-API-SECRET", c.apiSecret)

	return c.httpClient.Do(req)
}

// Send can be used to send a notification, using `POST notification` under the hood
// https://documentation.engagespot.co/docs/rest-api#tag/Notifications/paths/~1v3~1notifications/post
func (c *client) Send(n *notification) (*http.Response, error) {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(n)

	req, err := http.NewRequest("POST", ENDPOINT+"notifications", b)
	if err != nil {
		return nil, err
	}

	return c.call(req)
}

// Connect can be used to activate a user account without the need to manually login using application.
// This is helpful for sending notifications before user's first login. Beware that this will mark the
// user as active. uses sdk/notifications behind the scenes
// https://documentation.engagespot.co/docs/rest-api#tag/Notifications/paths/~1v3~1notifications/post
func (c *client) Connect(userId string) (*http.Response, error) {
	req, err := http.NewRequest("POST", ENDPOINT+"sdk/connect", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-ENGAGESPOT-USER-ID", userId)
	req.Header.Add("X-ENGAGESPOT-DEVICE-ID", DEVICE_TYPE)

	if c.config.enableHmac {
		req.Header.Add("X-ENGAGESPOT-USER-SIGNATURE", c.GenHmac(userId))
	}

	return c.call(req)
}

// GenHmac can be used to generate sha256 required if Hmac is enabled.
// Read more: https://documentation.engagespot.co/docs/HMAC-authentication/enabling-HMAC-authentication
func (c *client) GenHmac(userId string) string {
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(userId))
	return hex.EncodeToString(h.Sum(nil))
}
