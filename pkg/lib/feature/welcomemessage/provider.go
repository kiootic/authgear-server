package welcomemessage

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type Provider struct {
	Context               context.Context
	LocalizationConfig    *config.LocalizationConfig
	MetadataConfiguration config.AppMetadata
	MessagingConfig       *config.MessagingConfig
	WelcomeMessageConfig  *config.WelcomeMessageConfig
	TemplateEngine        *template.Engine
	TaskQueue             task.Queue
}

func (p *Provider) send(emails []string) (err error) {
	if !p.WelcomeMessageConfig.Enabled {
		return
	}

	if p.WelcomeMessageConfig.Destination == config.WelcomeMessageDestinationFirst {
		if len(emails) > 1 {
			emails = emails[0:1]
		}
	}

	if len(emails) <= 0 {
		return
	}

	var emailMessages []mail.SendOptions
	for _, email := range emails {
		data := map[string]interface{}{
			"email": email,
		}

		preferredLanguageTags := intl.GetPreferredLanguageTags(p.Context)
		data["appname"] = intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(p.LocalizationConfig.FallbackLanguage), p.MetadataConfiguration, "app_name")

		renderCtx := &template.RenderContext{
			PreferredLanguageTags: preferredLanguageTags,
		}

		var textBody string
		textBody, err = p.TemplateEngine.Render(
			renderCtx,
			TemplateItemTypeWelcomeEmailTXT,
			data,
		)
		if err != nil {
			return
		}

		var htmlBody string
		htmlBody, err = p.TemplateEngine.Render(
			renderCtx,
			TemplateItemTypeWelcomeEmailHTML,
			data,
		)
		if err != nil {
			return
		}

		emailMessages = append(emailMessages, mail.SendOptions{
			MessageConfig: config.NewEmailMessageConfig(
				p.MessagingConfig.DefaultEmailMessage,
				p.WelcomeMessageConfig.EmailMessage,
			),
			Recipient: email,
			TextBody:  textBody,
			HTMLBody:  htmlBody,
		})
	}

	p.TaskQueue.Enqueue(&tasks.SendMessagesParam{
		EmailMessages: emailMessages,
	})

	return
}

func (p *Provider) SendToIdentityInfos(infos []*identity.Info) (err error) {
	var emails []string
	for _, info := range infos {
		if email, ok := info.Claims[identity.StandardClaimEmail].(string); ok {
			emails = append(emails, email)
		}
	}
	return p.send(emails)
}