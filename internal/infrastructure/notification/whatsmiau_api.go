package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	appservice "github.com/lambdavault/api/internal/application/service"
)

type WhatsmiauAPIConfig struct {
	BaseURL      string
	APIKey       string
	InstanceName string
	DefaultPhone string
	Timeout      time.Duration
}

type WhatsmiauAPINotifier struct {
	baseURL      string
	apiKey       string
	instanceName string
	defaultPhone string
	client       *http.Client
}

func NewWhatsmiauAPINotifier(cfg WhatsmiauAPIConfig) *WhatsmiauAPINotifier {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &WhatsmiauAPINotifier{
		baseURL:      strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		apiKey:       strings.TrimSpace(cfg.APIKey),
		instanceName: strings.TrimSpace(cfg.InstanceName),
		defaultPhone: normalizeWhatsmiauNumber(cfg.DefaultPhone),
		client:       &http.Client{Timeout: timeout},
	}
}

func (n *WhatsmiauAPINotifier) Enabled() bool {
	return n != nil && n.baseURL != "" && n.apiKey != "" && n.instanceName != ""
}

func (n *WhatsmiauAPINotifier) NotifyPasswordCreated(ctx context.Context, event appservice.PasswordCreatedEvent) error {
	message := formatPasswordCreatedMessage(event)
	return n.sendToMany(ctx, candidateRecipients(event.RecipientPhones, []string{event.OwnerPhone}, []string{n.defaultPhone}), message)
}

func (n *WhatsmiauAPINotifier) NotifyPasswordShared(ctx context.Context, event appservice.PasswordSharedEvent) error {
	message := formatPasswordSharedMessage(event)
	return n.sendToMany(ctx, candidateRecipients(event.RecipientPhones, []string{n.defaultPhone}), message)
}

func (n *WhatsmiauAPINotifier) NotifyGroupMemberAdded(ctx context.Context, event appservice.GroupMemberAddedEvent) error {
	message := formatGroupMemberAddedMessage(event)
	return n.sendToMany(ctx, candidateRecipients(event.RecipientPhones, []string{event.MemberPhone}, []string{n.defaultPhone}), message)
}

func (n *WhatsmiauAPINotifier) sendToMany(ctx context.Context, numbers []string, text string) error {
	if !n.Enabled() || len(numbers) == 0 || strings.TrimSpace(text) == "" {
		return nil
	}
	var failed []string
	for _, number := range numbers {
		if err := n.sendText(ctx, number, text); err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", number, err))
		}
	}
	if len(failed) > 0 {
		sort.Strings(failed)
		return fmt.Errorf("whatsmiau send failures: %s", strings.Join(failed, "; "))
	}
	return nil
}

func (n *WhatsmiauAPINotifier) sendText(ctx context.Context, number, text string) error {
	number = normalizeWhatsmiauNumber(number)
	if number == "" {
		return nil
	}

	body, err := json.Marshal(map[string]any{
		"number": number,
		"text":   text,
	})
	if err != nil {
		return fmt.Errorf("marshal whatsmiau request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/message/sendText/%s", n.baseURL, url.PathEscape(n.instanceName))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build whatsmiau request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("apikey", n.apiKey)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("send whatsmiau request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	return nil
}

func candidateRecipients(groups ...[]string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0)
	for _, group := range groups {
		for _, value := range group {
			number := normalizeWhatsmiauNumber(value)
			if number == "" {
				continue
			}
			if _, ok := seen[number]; ok {
				continue
			}
			seen[number] = struct{}{}
			out = append(out, number)
		}
	}
	sort.Strings(out)
	return out
}

var nonDigits = regexp.MustCompile(`[^0-9]`)

func normalizeWhatsmiauNumber(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return nonDigits.ReplaceAllString(raw, "")
}

func formatPasswordCreatedMessage(event appservice.PasswordCreatedEvent) string {
	lines := []string{
		"🔐 LambdaVault",
		"Se creó una nueva contraseña.",
		"Sitio: " + safe(event.SiteName),
		"Usuario: " + safe(event.Username),
	}
	if strings.TrimSpace(event.Category) != "" {
		lines = append(lines, "Categoría: "+event.Category)
	}
	lines = append(lines, "Fecha: "+event.OccurredAt.Format(time.RFC3339))
	return strings.Join(lines, "\n")
}

func formatPasswordSharedMessage(event appservice.PasswordSharedEvent) string {
	lines := []string{
		"🤝 LambdaVault",
		"Te compartieron una contraseña.",
		"Grupo: " + safe(event.GroupName),
		"Sitio: " + safe(event.SiteName),
		"Usuario: " + safe(event.Username),
		"Compartido por: " + safe(event.SharedByEmail),
	}
	if strings.TrimSpace(event.Category) != "" {
		lines = append(lines, "Categoría: "+event.Category)
	}
	lines = append(lines, "Fecha: "+event.OccurredAt.Format(time.RFC3339))
	return strings.Join(lines, "\n")
}

func formatGroupMemberAddedMessage(event appservice.GroupMemberAddedEvent) string {
	lines := []string{
		"👥 LambdaVault",
		"Te agregaron a un grupo de contraseñas.",
		"Grupo: " + safe(event.GroupName),
		"Rol: " + safe(event.MemberRole),
		"Agregado por: " + safe(event.InvitedByEmail),
	}
	lines = append(lines, "Fecha: "+event.OccurredAt.Format(time.RFC3339))
	return strings.Join(lines, "\n")
}

func safe(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}
