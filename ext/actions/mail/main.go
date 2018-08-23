// Package mailaction provides an action to send an email via SMTP
package mailaction

import (
	"strconv"
	"fmt"
	"net/smtp"

	"gopkg.in/yaml.v2"

	"github.com/ObjectifLibre/csf/actions"
)

func init() {
	actions.RegisterActionModule("mail", &mailActionModuleImplementation{})
}

var _ actions.ActionModuleInterface = mailActionModuleImplementation{}

type mailActionModuleImplementation struct {}

type smtpConfig struct {
	Port int `yaml:"port"`
	Server string `yaml:"server"`
	User string `yaml:"user"`
	Password string `yaml:"password"`
}

var config = smtpConfig{}

func (mail mailActionModuleImplementation) Actions() (map[string][]actions.ArgType,map[string][]actions.ArgType) {
	in := map[string][]actions.ArgType{
		"send_mail": {{T: "string", N: "to"},
			{T: "string", N: "subject"},
			{T: "string", N: "content"},
		}}
	out := map[string][]actions.ArgType{
		"send_mail": {}}
	return in, out
}

func sendMail(data map[string]interface{}) error {
	to, ok := data["to"].(string)
	if !ok {
		return fmt.Errorf("Bad data, expecting 'to' as string")
	}
	subject, ok := data["subject"].(string)
	if !ok {
		return fmt.Errorf("Bad data, expecting 'subject' as string")
	}
	content, ok := data["content"].(string)
	if !ok {
		return fmt.Errorf("Bad data, expecting 'content' as string")
	}
	msg := "From: " + config.User + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		content

	auth := smtp.PlainAuth("", config.User, config.Password, config.Server)
	err := smtp.SendMail(config.Server + ":" + strconv.Itoa(config.Port), auth,
		        config.User, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("Could not send mail: %s", err)
	}
	return nil
}

func mailActionHandler(action string, data map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "send_mail":
 		return nil, sendMail(data)
	default:
		return nil, fmt.Errorf("No such action: " + action)
	}
}

func (mail mailActionModuleImplementation) Setup(rawcfg []byte) error {
	if err := yaml.Unmarshal(rawcfg, &config); err != nil {
		return fmt.Errorf("Error in config: %s", err)
	}
	return nil
}

func (mail mailActionModuleImplementation) ActionHandler() func(string, map[string]interface{}) (map[string]interface{}, error) {
	return mailActionHandler
}
