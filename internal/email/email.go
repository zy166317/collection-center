package email

import (
	"collection-center/internal/logger"
	"collection-center/service/db/dao"
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

func SendEmail(order *dao.Orders) {
	subject := "Please check your order"
	body := fmt.Sprintf(`
                <section style="padding:20px;">
                    <h5 style="text-align: center;">Your order ID is %d</h5>
                    <p style="text-align: center;">%v %v To %v %v</p>
                    <hr />
                    <table>
                        <tbody>
                            <tr>
                                <td>order Type</td>
                                <td>%v</td>
                            </tr>
                            <tr>
                                <td>order Status</td>
                                <td>PENDING</td>
                            </tr>
                            <tr>
                                <td>Send</td>
                                <td>%v %v</td>
                            </tr>
                            <tr>
                                <td>Receive</td>
                                <td>%v %v</td>
                            </tr>
                            <tr>
                                <td>Receive Address</td>
                                <td>%v</td>
                            </tr>
                        </tbody>
                    </table>
                    <p>To make an exchange, send %v %v to the address within 30 minutes:</p>
                    <h5 style="text-align: center;">%v</h5>
                </section>
            `,
		order.Id,
		order.OriginalTokenAmount,
		order.OriginalToken,
		order.TargetTokenAmount,
		order.TargetToken,
		order.Mode,
		order.OriginalTokenAmount,
		order.OriginalToken,
		order.TargetTokenAmount,
		order.TargetToken,
		order.UserReceiveAddress,
		order.OriginalTokenAmount,
		order.OriginalToken,
		order.WeReceiveAddress,
	)

	m := gomail.NewMessage()
	// 设置电子邮件的基本信息
	m.SetHeader("From", conn.User)
	m.SetHeader("To", order.Email)
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
