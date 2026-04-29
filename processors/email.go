package processors

type EmailData struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Html    string `json:"html"`
}

func SendEmail(data map[string]any) {
	// verify data is EmailData
	// verify From is email format
	// verify to is email format
	// emailapi.SendEmail()
}
