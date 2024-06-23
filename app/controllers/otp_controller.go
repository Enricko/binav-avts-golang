package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/smtp"
	"os"

	"github.com/gin-gonic/gin"
)

type OtpController struct {
	OTP string
}

type OTPResponse struct {
	Message string `json:"message"`
}

func NewOtpController() *OtpController {
	return &OtpController{
		// Inject services
	}
}

const (
	smtpServer           = "smtp.gmail.com"
	smtpPort             = 587
	senderEmailEnvVar    = "fadillarizky294@gmail.com"
	senderPasswordEnvVar = "#####"
)

func GetSenderEmail() (string, error) {
	email := os.Getenv(senderEmailEnvVar)
	if email == "" {
		return "", fmt.Errorf("missing environment variable: %s", senderEmailEnvVar)
	}
	return email, nil
}

func GetSenderPassword() (string, error) {
	password := os.Getenv(senderPasswordEnvVar)
	if password == "" {
		return "", fmt.Errorf("missing environment variable: %s", senderPasswordEnvVar)
	}
	return password, nil
}

func GenerateOTP() string {
	// Implement your logic to generate a random OTP code (e.g., using math/rand)
	return "123456" // Replace with actual generated code
}

const emailTemplate = `
<!DOCTYPE html>
<html>
<body>
  <h2>Reset Password OTP</h2>
  <p>Your OTP code to reset your password is: {{.OTP}}</p>
</body>
</html>
`

// Parse the template
var t *template.Template

func init() {
	var err error
	t, err = template.New("Foo").Parse(emailTemplate)
	if err != nil {
		panic(err)
	}
}

func sendEmail(recipientEmail string, otp string) error {
	senderEmail, err := GetSenderEmail()
	if err != nil {
		return err
	}
	senderPassword, err := GetSenderPassword()
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpServer)

	// Render the template with data
	var htmlBody bytes.Buffer
	err = t.Execute(&htmlBody, OtpController{OTP: otp})
	if err != nil {
		return err
	}

	msg := []byte(fmt.Sprintf("Subject: Reset Password OTP\n\n%s\n", htmlBody.String()))

	err = smtp.SendMail(fmt.Sprintf("%s:%d", smtpServer, smtpPort), auth, senderEmail, []string{recipientEmail}, msg)
	if err != nil {
		return err
	}

	fmt.Println("Email sent successfully!")
	return nil
}

func (r *OtpController) InsertOtp(c *gin.Context) {
	// ... (Parse email from request)

	// if http.Request.Method != http.MethodPost {
	// 	c.JSON(http.StatusMethodNotAllowed, gin.H{"message": "invalid request method"})
	// 	return
	// }

	otp := GenerateOTP()

	err := sendEmail("ixb.11.fadillarizky@gmail.com", otp)
	if err != nil {
		// Handle error (e.g., log the error, return error message)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Success"})

	// ... (Display success message or redirect to confirmation page)
}
