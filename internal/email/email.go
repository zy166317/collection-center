package email

import (
	"collection-center/internal/logger"
	"fmt"
	"gopkg.in/gomail.v2"
)

type EmailConfig struct {
	Host string
	Port int
	User string
	Pass string
}

var conn EmailConfig

func InitEmail(c *EmailConfig) {
	conn.Host = c.Host
	conn.Port = c.Port
	conn.User = c.User
	conn.Pass = c.Pass
}

type BaseMailInfo struct {
	MailAddress      string //邮箱地址
	VerificationCode string //验证码
}

//func SendEmail(from string, to string, subject string, body string) {
//	m := gomail.NewMessage()
//	// 设置电子邮件的基本信息
//	m.SetHeader("From", from)
//	m.SetHeader("To", to)
//	m.SetHeader("Subject", subject)
//	m.SetBody("text/plain", body)
//	// 设置SMTP服务器的详细信息
//	d := gomail.NewDialer(conn.Host, conn.Port, conn.User, conn.Pass) // 设置邮件正文
//	// 发送电子邮件
//	if err := d.DialAndSend(m); err != nil {
//		logger.Fatal(err)
//	}
//	logger.Info("Email sent!")
//}

func SendEmail(mail *BaseMailInfo) {
	subject := "Please verify your email address"
	body := fmt.Sprintf(`
               <section style="padding:20px;">
                   <h5 style="text-align: center;"> %s</h5>
                   <p style="text-align: center;">Verification code:%s</p>
                   <hr />
                   <table>
                   </table>
                   <p></p>
                   <h5 style="text-align: center;">The verification code is valid for 30 minutes</h5>
               </section>
           `,
		mail.MailAddress,
		mail.VerificationCode,
	)

	m := gomail.NewMessage()
	// 设置电子邮件的基本信息
	m.SetHeader("From", conn.User)
	m.SetHeader("To", mail.MailAddress)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	// 设置SMTP服务器的详细信息
	d := gomail.NewDialer(conn.Host, conn.Port, conn.User, conn.Pass) // 设置邮件正文
	// 发送电子邮件
	if err := d.DialAndSend(m); err != nil {
		logger.Errorf("Send email error:%s", err)
		return
	}
	logger.Info("Email sent!")
}
