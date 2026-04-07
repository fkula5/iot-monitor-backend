package services

import (
	"testing"
)

func TestNewMailer(t *testing.T) {
	mailer := NewMailer("smtp.example.com", 587, "user", "pass", "from@example.com", "http://localhost:3000")
	if mailer == nil {
		t.Fatal("Expected NewMailer to return a non-nil object")
	}
	if mailer.from != "from@example.com" {
		t.Errorf("Expected from to be from@example.com, got %s", mailer.from)
	}
	if mailer.frontendURL != "http://localhost:3000" {
		t.Errorf("Expected frontendURL to be http://localhost:3000, got %s", mailer.frontendURL)
	}
}

func TestSendResetPasswordEmail(t *testing.T) {
	mailer := NewMailer("localhost", 1025, "", "", "from@example.com", "http://localhost:3000")
	err := mailer.SendResetPasswordEmail("to@example.com", "mytoken")
	// Since no server is running at localhost:1025, it should return an error
	if err == nil {
		t.Error("Expected an error since no SMTP server is running")
	}
}
