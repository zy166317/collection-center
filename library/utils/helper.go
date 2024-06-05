package utils

import (
	"bytes"
	"collection-center/contract/constant"
	"collection-center/internal/logger"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/image/font"
	"golang.org/x/xerrors"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	face font.Face
)

func GenerateMd5(string string) string {
	h := md5.New()
	h.Write([]byte(string))
	return hex.EncodeToString(h.Sum(nil))
}

func DecodeBase64Image(src string) (*[]byte, string) {
	data := strings.Split(src, ";")
	suffix := "png"
	if strings.Contains(data[0], "jpeg") {
		suffix = "jpg"
	}
	decodeBytes, err := base64.StdEncoding.DecodeString(strings.Split(data[1], ",")[1])
	if err != nil {
		return nil, ""
	}
	return &decodeBytes, suffix
}

// 根据map的key排序，返回排好序的key slice
func SortByKey(m map[string]string) []string {
	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

func VerifyPhoneFormat(phone string) bool {
	reg := `^1([38][0-9]|14[57]|5[^4])\d{8}$`
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(phone)
}

func VerifyIdcardFormat(idcard string) bool {
	var reg string
	if len(idcard) == 18 {
		reg = `^[1-9]\d{5}(18|19|([23]\d))\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$`
	} else {
		reg = `^[1-9]\d{5}\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{2}[0-9Xx]$`
	}
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(idcard)
}

func Retry(retries int, fn func() error) error {
	if err := fn(); err != nil {
		retries--
		if retries <= 0 {
			return err
		}
		// preventing thundering herd problem (https://en.wikipedia.org/wiki/Thundering_herd_problem)
		time.Sleep(time.Millisecond * 100)
		return Retry(retries, fn)
	}
	return nil
}

func RunWithRecovery(fn func() error) error {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("RunWithRecovery: panic running job: %v\n%s", r, buf)
		}
	}()
	return fn()
}

func GenerateCaptcha() string {
	return fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
}

func Float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func ParseTimeStampToTime(timeStamp int64) time.Time {
	numberLength := len(strconv.FormatInt(timeStamp, 10))
	if numberLength >= 10 {
		timeStamp = timeStamp / Power(10, numberLength-10)
		return time.Unix(timeStamp, 0)
	} else {
		timeStamp = timeStamp * int64((10-numberLength)*10)
		return time.Unix(timeStamp, 0)
	}
}

func Power(base int64, exponent int) int64 {
	if exponent <= 0 {
		return 0
	}
	ans := int64(1)
	for exponent != 0 {
		ans *= base
		exponent--
	}
	return ans
}

func GetRandNum(len int) string {
	r := make([]string, len)
	for i := 0; i < len; i++ {
		r = append(r, strconv.Itoa(rand.Intn(9)))
	}
	return strings.Join(r, "")
}

func MakeDir(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(path, os.ModePerm)
		}
		return err
	}
	return nil

}

func FindInString(target string, elements *[]string) int {
	if elements == nil {
		return -1
	}
	for index, element := range *elements {
		if target == element {
			return index
		}
	}
	return -1
}

func CheckPhoneNumber(mobile string, regionCode string) bool {
	switch regionCode {
	case "+86":
		match, err := regexp.MatchString("^1[0-9]{10}$", mobile)
		if err != nil {
			return false
		}
		if !match {
			return false
		}
		return true
	case "+886":
		return true
	case "+852":
		return true
	default:
		fmt.Println("not supported region")
		return false
	}
}

// 产生6位数随机验证码
func VerifyCode() string {
	code := strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(899999) + 100000)
	return code
}

func GetRandStr(n int) (randStr string) {
	chars := "ABCDEFGHIJKMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"
	charsLen := len(chars)
	if n > 10 {
		n = 10
	}

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		randIndex := rand.Intn(charsLen)
		randStr += chars[randIndex : randIndex+1]
	}
	return randStr
}

func DrawCaptcha(width, height int, text string, path string) string {
	textLen := len(text)
	dc := gg.NewContext(width, height)
	bgR, bgG, bgB, bgA := getRandColorRange(240, 255)
	dc.SetRGBA255(bgR, bgG, bgB, bgA)
	dc.Clear()

	// 干扰线
	for i := 0; i < 20; i++ {
		x1, y1 := getRandPos(width, height)
		x2, y2 := getRandPos(width, height)
		r, g, b, a := getRandColor(255)
		w := float64(rand.Intn(3) + 1)
		dc.SetRGBA255(r, g, b, a)
		dc.SetLineWidth(w)
		dc.DrawLine(x1, y1, x2, y2)
		dc.Stroke()
	}

	fontSize := float64(height/2) + 5
	if face == nil {
		face = loadFontFace(fontSize, path)
	}
	dc.SetFontFace(face)
	for i := 0; i < len(text); i++ {
		r, g, b, _ := getRandColor(100)
		dc.SetRGBA255(r, g, b, 255)
		fontPosX := float64(width/textLen*i) + fontSize*0.2
		fontPosY := float64(height/2) + fontSize*0.5*(rand.Float64()-1)

		writeText(dc, text[i:i+1], float64(fontPosX), fontPosY)
	}

	buffer := bytes.NewBuffer(nil)
	_ = dc.EncodePNG(buffer)
	b := buffer.Bytes()
	sourceString := "data:image/png;base64," + base64.StdEncoding.EncodeToString(b)
	return sourceString
}

// 渲染文字
func writeText(dc *gg.Context, text string, x, y float64) {
	xfload := 5 - rand.Float64()*10 + x
	yfload := 5 - rand.Float64()*10 + y

	radians := 40 - rand.Float64()*80
	dc.RotateAbout(gg.Radians(radians), x, y)
	dc.DrawStringAnchored(text, xfload, yfload, 0.2, 0.5)
	dc.RotateAbout(-1*gg.Radians(radians), x, y)
	dc.Stroke()
}

// 随机坐标
func getRandPos(width, height int) (x float64, y float64) {
	x = rand.Float64() * float64(width)
	y = rand.Float64() * float64(height)
	return x, y
}

// 随机颜色
func getRandColor(maxColor int) (r, g, b, a int) {
	r = int(uint8(rand.Intn(maxColor)))
	g = int(uint8(rand.Intn(maxColor)))
	b = int(uint8(rand.Intn(maxColor)))
	a = int(uint8(rand.Intn(255)))
	return r, g, b, a
}

// 随机颜色范围
func getRandColorRange(miniColor, maxColor int) (r, g, b, a int) {
	if miniColor > maxColor {
		miniColor = 0
		maxColor = 255
	}
	r = int(uint8(rand.Intn(maxColor-miniColor) + miniColor))
	g = int(uint8(rand.Intn(maxColor-miniColor) + miniColor))
	b = int(uint8(rand.Intn(maxColor-miniColor) + miniColor))
	a = int(uint8(rand.Intn(maxColor-miniColor) + miniColor))
	return r, g, b, a
}

// 加载字体
func loadFontFace(points float64, path string) font.Face {
	// 这里是将字体TTF文件转换成了 byte 数据保存成了一个 go 文件 文件较大可以到附录下
	// 通过truetype.Parse可以将 byte 类型的数据转换成TTF字体类型
	ttf, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	f, err := truetype.Parse(ttf)

	if err != nil {
		panic(err)
	}
	face := truetype.NewFace(f, &truetype.Options{
		Size: points,
	})
	return face
}

// 强制转换 string -> int
func MustAtoI(s string) int {
	num, _ := strconv.Atoi(s)
	return num
}

func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// 将元转换为分
var oneHundred, _ = decimal.NewFromString("100")

func ConvertStringYuan2Fen(yuan string) int64 {
	source, err := decimal.NewFromString(yuan)
	if err != nil {
		fmt.Print("转换元到分异常", err.Error())
		return 0
	}
	return source.Mul(oneHundred).IntPart()
}

func ConvertStringFen2Fen(fen string) int64 {
	source, err := decimal.NewFromString(fen)
	if err != nil {
		fmt.Print("转换元到分异常", err.Error())
		return 0
	}
	return source.IntPart()
}

func ConvertInt64Fen2Yuan(fen int64) string {
	source := decimal.New(fen, 0)
	res := source.DivRound(oneHundred, 2)
	return res.String()
}

func Convert2ProtoTime(time time.Time) *timestamppb.Timestamp {
	if time.IsZero() {
		return nil
	}
	reT := timestamppb.New(time)
	return reT
}
func ConvertPointer2ProtoTime(time *time.Time) *timestamppb.Timestamp {
	if time == nil || time.IsZero() {
		return nil
	}
	reT := timestamppb.New(*time)
	return reT
}

// 将map按key字典排序后，value转成string拼接后md5
func GenMapValueMd5(params map[string]interface{}) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vals := ""
	for _, k := range keys {
		vals = vals + Strval(params[k])
	}
	return GenerateMd5(vals)
}

func Strval(value interface{}) string {
	// interface 转 string
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

func Md5Sum(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("空的data")
	}
	hasher := md5.New()
	hasher.Write(data)
	return hasher.Sum(nil), nil
}

func MustMd5Sum(data []byte) string {
	rst, err := Md5Sum(data)
	if err != nil {
		logger.Error(err)
	}
	return string(rst)
}

func CheckToken(token string) error {
	tokens := []string{
		"ETH",
		"USDT",
		"BTC",
	}
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == token {
			return nil
		}
	}

	return xerrors.New("Invalid token name")
}

func VerifyEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// BtcSatToB btc satoshi to btc 聪 转 btc
func BtcSatToB(sat int64) *big.Float {
	decimals8, _ := StrToBigFloat(constant.DECIMALS_EIGHT)

	return new(big.Float).Quo(
		new(big.Float).SetInt(big.NewInt(sat)),
		decimals8,
	)
}

func RangeRandom(min, max int) (number int) {
	//创建随机种子
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	number = r.Intn(max-min) + min
	return number
}
