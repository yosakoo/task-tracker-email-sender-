package models

type EmailMessage struct {
    Subject string `json:"subject"`
    Body    string `json:"body"`
    To      string `json:"to"`
}
