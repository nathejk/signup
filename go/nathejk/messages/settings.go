package messages

import (
	"time"

	"nathejk.dk/nathejk/types"
)

type NathejkPatruljeSignupOpened struct {
	MaxSeatCount int
}

type NathejkPatruljeSignupClosed struct {
}

type NathejkKlanSignupOpened struct {
	MaxSeatCount int
}

type NathejkKlanSignupClosed struct {
}

type NathejkKlanSignupStartSpecified struct {
	Time *time.Time
}

type NathejkMailTemplateUpdated struct {
	Slug     types.Slug
	Subject  string
	Template string
}

type NathejkMailTemplateDeleted struct {
	Slug types.Slug
}
