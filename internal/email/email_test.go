package email

import (
	"testing"
)

func TestSendEmail(t *testing.T) {
	conn.Host = "smtp.gmail.com"
	conn.Port = 465
	conn.User = "orcayihaoji@gmail.com"
	conn.Pass = "ifgmqpidheqgpctm"
	//SendEmail(conn.User, "clydekuo6@gmail.com", "Test Email", "Test Body")
}
