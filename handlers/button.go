package handlers

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
)

type btn struct {
	Label    string                    `json:"label"`
	Style    discordgo.ButtonStyle     `json:"style"`
	Disabled bool                      `json:"disabled,omitempty"`
	Emoji    *discordgo.ComponentEmoji `json:"emoji,omitempty"`
	URL      string                    `json:"url,omitempty"`
	CustomID string                    `json:"custom_id,omitempty"`
}

func (b btn) Type() discordgo.ComponentType { return discordgo.ButtonComponent }

func (b btn) MarshalJSON() ([]byte, error) {
	if b.Style == 0 {
		b.Style = discordgo.PrimaryButton
	}
	type alias btn
	return json.Marshal(struct {
		alias
		Type discordgo.ComponentType `json:"type"`
	}{alias: alias(b), Type: discordgo.ButtonComponent})
}

func mkBtn(label string, style discordgo.ButtonStyle, customID string) discordgo.MessageComponent {
	return btn{Label: label, Style: style, CustomID: customID}
}
