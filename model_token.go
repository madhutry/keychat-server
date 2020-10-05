package main

type Token struct {
	Token          string `json:"token,omitempty"`
	AvatarUrl      string `json:"avatarUrl,omitempty"`
	WelcomeMessage string `json:welcomeMessage,omitempty"`
}
