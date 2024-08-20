package helper

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"golang-app/app/models"
	"html/template"
	"net/smtp"
	"os"
)

// Helper function to generate OTP
func GenerateOTP(length int) (string, error) {
	const otpChars = "1234567890"
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer), nil
}
func SendOTPEmail(user models.User, otp string) error {
	from := os.Getenv("EMAIL_FROM")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	companyName := os.Getenv("COMPANY_NAME")
	baseURL := os.Getenv("BASE_URL") // Add this to your .env file

	// Construct the logo URL
	logoUrl := baseURL + "/public/assets/logo_transparent.png"

	to := []string{user.Email}

	// HTML email template
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background-color: #f4f4f4; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 20px auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        .header { background-color: #ffffff; padding: 20px; text-align: center; border-bottom: 2px solid #0056b3; }
        .logo-container { background-color: #ffffff; padding: 20px; display: inline-block; border-radius: 50%; }
        .logo { max-width: 150px; height: auto; }
        .title { color: #0056b3; margin-top: 20px; }
        .content { padding: 20px; }
        .code-box { background-color: #e9ecef; padding: 15px; margin: 20px 0; text-align: center; border-radius: 4px; }
        .code { font-size: 32px; font-weight: bold; color: #0056b3; letter-spacing: 5px; }
        .warning { background-color: #fff3cd; border: 1px solid #ffeeba; color: #856404; padding: 10px; margin-top: 20px; border-radius: 4px; }
        .footer { font-size: 12px; color: #6c757d; margin-top: 30px; text-align: center; border-top: 1px solid #dee2e6; padding-top: 20px; }
        @media only screen and (max-width: 600px) {
            .container { width: 100%; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo-container">
                <img src="{{.LogoUrl}}" alt="{{.CompanyName}} Logo" class="logo">
            </div>
            <h1 class="title">Password Reset</h1>
        </div>
        <div class="content">
            <p>Hello {{.Name}},</p>
            <p>We received a request to reset the password for your account. Here is your One-Time Password (OTP) to proceed with the password reset:</p>
            <div class="code-box">
                <div class="code">{{.OTP}}</div>
            </div>
            <p>This OTP will expire in 15 minutes for security reasons.</p>
            <div class="warning">
                <strong>Security Notice:</strong>
                <p>If you didn't request this password reset, someone might be trying to access your account. We recommend that you secure your account immediately by changing your password.</p>
            </div>
            <p>The OTP contained in this email is required to reset your password. Do not share this code with anyone.</p>
            <div class="footer">
                <p>This is an automated email from {{.CompanyName}}. Please do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>
`

	// Prepare email data
	data := struct {
		Name        string
		Email       string
		OTP         string
		CompanyName string
		LogoUrl     string
	}{
		Name:        user.Name,
		Email:       user.Email,
		OTP:         otp,
		CompanyName: companyName,
		LogoUrl:     logoUrl,
	}

	// Parse and execute the template
	t, err := template.New("emailTemplate").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	// Compose the email
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: Password Reset Request - " + companyName + "\n"
	msg := []byte(subject + mime + body.String())

	// Send the email
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
