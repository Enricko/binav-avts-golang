package helper

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"golang-app/app/models"
	"html/template"
	"net/smtp"
	"os"
	"time"
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

func SendEmail(user models.User) error {
	from := os.Getenv("EMAIL_FROM")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	companyName := os.Getenv("COMPANY_NAME")
	baseURL := os.Getenv("BASE_URL") // Add this to your .env file

	// Check if all required environment variables are set
	if from == "" || password == "" || smtpHost == "" || smtpPort == "" || companyName == "" || baseURL == "" {
		return fmt.Errorf("missing required environment variables")
	}

	logoUrl := baseURL + "/public/assets/logo_transparent.png"
	to := []string{user.Email}

	// HTML email template
	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User Details Confirmation</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background-color: #f4f4f4; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 20px auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        .header { background-color: #0056b3; padding: 20px; text-align: center; }
        .logo { max-width: 150px; height: auto; }
        .title { color: #ffffff; margin-top: 20px; font-size: 24px; }
        .content { padding: 30px; }
        .user-details { background-color: #e9ecef; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .detail-item { margin-bottom: 10px; }
        .detail-label { font-weight: bold; color: #0056b3; }
        .button { display: inline-block; padding: 10px 20px; background-color: #0056b3; color: #ffffff !important; text-decoration: none; border-radius: 4px; margin-top: 20px; font-weight: bold; }
        .footer { font-size: 12px; color: #6c757d; margin-top: 30px; text-align: center; border-top: 1px solid #dee2e6; padding-top: 20px; }
        @media only screen and (max-width: 600px) {
            .container { width: 100%; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <img src="{{.LogoUrl}}" alt="{{.CompanyName}} Logo" class="logo">
            <h1 class="title">User Details Confirmation</h1>
        </div>
        <div class="content">
            <p>Dear {{.Name}},</p>
            <p>Thank you for registering with {{.CompanyName}}. We have successfully received your user details. Please review the information below to ensure its accuracy:</p>
            
            <div class="user-details">
                <div class="detail-item">
                    <span class="detail-label">Name:</span> {{.Name}}
                </div>
                <div class="detail-item">
                    <span class="detail-label">Email:</span> {{.Email}}
                </div>
                <!-- Add more user details as needed -->
            </div>
            
            <p>If you need to make any changes to your information or have any questions, please don't hesitate to contact our support team.</p>
            
            <p>You can access your account and explore our services by visiting our website:</p>
            
            <a href="{{.BaseUrl}}" class="button">Visit AVTS Website</a>
            
            <p>Thank you for choosing {{.CompanyName}}. We look forward to serving you.</p>
            
            <p>Best regards,<br>The {{.CompanyName}} Team</p>
        </div>
        <div class="footer">
            <p>This email was sent to {{.Email}}. If you did not create an account with {{.CompanyName}}, please disregard this email.</p>
            <p>&copy; {{.CurrentYear}} {{.CompanyName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	// Prepare email data
	data := struct {
		Name        string
		Email       string
		CompanyName string
		LogoUrl     string
		BaseUrl     string
		CurrentYear int
	}{
		Name:        user.Name,
		Email:       user.Email,
		CompanyName: companyName,
		LogoUrl:     logoUrl,
		BaseUrl:     baseURL,
		CurrentYear: time.Now().Year(),
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
	subject := "Subject: Email Confirmation - " + companyName + "\n"
	msg := []byte(subject + mime + body.String())

	// Send the email
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
