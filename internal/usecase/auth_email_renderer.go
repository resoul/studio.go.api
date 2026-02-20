package usecase

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"
)

//go:embed templates/email/*/*.tmpl
var emailTemplatesFS embed.FS

func renderLocalizedEmail(locale, templateName string, data emailTemplateData) (renderedEmail, error) {
	lang := normalizeLocale(locale)

	subject, err := renderTextTemplate(templatePath(lang, templateName, "subject"), data)
	if err != nil && lang != "en" {
		subject, err = renderTextTemplate(templatePath("en", templateName, "subject"), data)
	}
	if err != nil {
		return renderedEmail{}, fmt.Errorf("failed to render subject for %s: %w", templateName, err)
	}
	data.Subject = strings.TrimSpace(subject)

	textBody, err := renderTextTemplate(templatePath(lang, templateName, "text"), data)
	if err != nil && lang != "en" {
		textBody, err = renderTextTemplate(templatePath("en", templateName, "text"), data)
	}
	if err != nil {
		return renderedEmail{}, fmt.Errorf("failed to render text for %s: %w", templateName, err)
	}

	htmlBody, err := renderHTMLTemplate(templatePath(lang, templateName, "html"), data)
	if err != nil && lang != "en" {
		htmlBody, err = renderHTMLTemplate(templatePath("en", templateName, "html"), data)
	}
	if err != nil {
		return renderedEmail{}, fmt.Errorf("failed to render html for %s: %w", templateName, err)
	}

	return renderedEmail{
		Subject:  data.Subject,
		TextBody: strings.TrimSpace(textBody),
		HTMLBody: htmlBody,
	}, nil
}

func renderTextTemplate(path string, data emailTemplateData) (string, error) {
	raw, err := emailTemplatesFS.ReadFile(path)
	if err != nil {
		return "", err
	}
	tpl, err := texttemplate.New(path).Parse(string(raw))
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}

func renderHTMLTemplate(path string, data emailTemplateData) (string, error) {
	raw, err := emailTemplatesFS.ReadFile(path)
	if err != nil {
		return "", err
	}
	tpl, err := htmltemplate.New(path).Parse(string(raw))
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}
