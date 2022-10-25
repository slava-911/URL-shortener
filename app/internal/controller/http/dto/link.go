package dto

import (
	"regexp"
	"strings"

	"github.com/slava-911/URL-shortener/internal/domain/entity"
)

type CreateLinkDTO struct {
	FullVersion string `json:"full_version" validate:"required,min=3,max=2000"`
	Description string `json:"description,omitempty"`
	UserID      string `json:"user_id" validate:"required"`
}

func ValidLink(link string) bool {
	re, err := regexp.Compile("^(http|https)://")
	if err != nil {
		return false
	}
	link = strings.TrimSpace(link)
	// Check if string matches the regex
	if re.MatchString(link) {
		return true
	}
	return false
}

func NewLink(d CreateLinkDTO) entity.Link {
	return entity.Link{
		FullVersion: d.FullVersion,
		Description: d.Description,
		UserID:      d.UserID,
	}
}

type UpdateLinkDTO struct {
	FullVersion *string `json:"full_version,omitempty"`
	Description *string `json:"description,omitempty"`
}
