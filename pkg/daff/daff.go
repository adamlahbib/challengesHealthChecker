package daff

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
)

//	challenges:
//		- name:
//			url:
//			request:
//				method:
//				headers:
//					-
//					-
//				cookies:
//					-
//					-
//				body:
//			response:
//				status:
//		- name:
//			url:
//			request:
//				method:
//				headers:
//					-
//					-
//				cookies:
//					-
//					-
//				body:
//			response:
//				status:

// Config stores the daff configurations
type Config struct {
	Challenges map[string]Challenge `yaml:"challenges,omitempty"`
}

// Challenge stores the details for a challenge
type Challenge struct {
	URL      string   `yaml:"url,omitempty"`
	Request  Request  `yaml:"request,omitempty"`
	Response Response `yaml:"response,omitempty"`
}

// Request stores details of a request
type Request struct {
	Method  string   `yaml:"method,omitempty"`
	Headers []string `yaml:"headers,omitempty"`
	Cookies []string `yaml:"cookies,omitempty"`
	Body    string   `yaml:"body,omitempty"`
}

// Response stores details of a request
type Response struct {
	Status int `yaml:"status,omitempty"`
}

const (
	headerDelimiter   = ":"
	prefix            = "/check"
	connectionRefused = "connection refused"
	responseMessage   = "@here Challenge `%v` is %s\n"

	up   = ":thumbsup:"
	down = ":thumbsdown:"
)

// New parses file and returns a new instance of daff config
func New(file string) (*Config, error) {
	config := &Config{}

	err := config.parseFile(file)
	if err != nil {
		log.Printf("Error parsing config file: %v\n", err)
		return nil, err
	}

	return config, nil
}

// Print pretty print the config
func (c *Config) Print() {
	bytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Printf("Error marshalling: %v\n", err)
		return
	}
	log.Printf("%s\n", string(bytes))
}

/*
// MessageCreate handles new message from channels that the bot has access to

	func (c *Config) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Check if the message is intended for the bot
		if !strings.HasPrefix(m.Content, prefix) {
			return
		}

		parts := strings.Split(m.Content, " ")
		if len(parts[1]) == 0 {
			return
		}

		challenge := parts[1]

		res, err := c.CheckSanity(challenge)
		if err != nil {
			log.Printf("Failed to check health: %v\n", err)
			// Do not return for `connection refused` error
			// `connection refused` indicates server is down
			if !strings.Contains(err.Error(), connectionRefused) {
				return
			}
		}

		var message string

		if res {
			message = fmt.Sprintf(responseMessage, challenge, up)
		} else {
			message = fmt.Sprintf(responseMessage, challenge, down)
		}

		log.Printf("Responding with message: %v\n", message)
		s.ChannelMessageSend(m.ChannelID, message)

}
*/
func (c *Config) Loop(s *discordgo.Session) {
	ticker := time.NewTicker(5 * time.Minute)

	for {
		select {
		case <-ticker.C:
			for name := range c.Challenges {
				res, err := c.CheckSanity(name)
				if err != nil {
					log.Printf("Failed to check health of %v: %v\n", name, err)
					// Do not return for `connection refused` error
					// `connection refused` indicates server is down
					if !strings.Contains(err.Error(), connectionRefused) {
						continue
					}
				}

				var message string
				if res {
					// up, do nothing, used to be message = fmt.Sprintf(responseMessage, name, up)
				} else {
					message = fmt.Sprintf(responseMessage, name, down)
					// the next two lines used to be after else.
					log.Printf("Sending message: %v\n", message)
					s.ChannelMessageSend("1089884138703700068", message)
				}

			}
		}
	}
}

// CheckSanity checks the health of a challenge
func (c *Config) CheckSanity(name string) (bool, error) {
	challenge, ok := c.Challenges[name]
	if !ok {
		log.Printf("Challenge configuration not found\n")
		return false, fmt.Errorf("challenge configuration not found")
	}

	log.Printf("Challenge configuration found: %+v\n", challenge)

	req, err := createRequest(challenge)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return false, err
	}

	log.Printf("Request created: %+v\n", req)

	res, err := sendRequest(req)
	if err != nil {
		log.Printf("Error sending request: %v\n", err)
		return false, err
	}

	log.Printf("Received response: %+v\n", res)

	return isResponseValid(res, &challenge.Response), nil
}

// parseFile parses the config file into a struct
func (c *Config) parseFile(file string) error {
	fileBytes, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("Error reading config file: %v\n", err)
		return err
	}

	log.Printf("Configuration file found\n")

	err = yaml.Unmarshal(fileBytes, c)
	if err != nil {
		log.Printf("Error unmarshalling config file: %v\n", err)
		return err
	}

	return nil
}
