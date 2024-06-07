package queue

import (
	"fmt"
	"net/http"
)

func SendMsg(verifyOrderNotify *VerifyOrderNotify) {
	url := fmt.Sprintf("http://%s:8080/test/ping", verifyOrderNotify.IPAddr)
	method := "GET"
	//jsonStr := fmt.Sprintf(`{"hash":%s,"amount":%s,"text":%s,"payAddr":%s}`, verifyOrderNotify.Hash, verifyOrderNotify.Amount, verifyOrderNotify.Text, verifyOrderNotify.PayAddr)
	//payload := strings.NewReader(jsonStr)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return
	}
	//req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		//失败重新放入队列
	}
}
