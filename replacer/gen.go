package replacer

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Pallinder/go-randomdata"
	xj "github.com/basgys/goxml2json"
	"github.com/jiwen624/uuid"
	"github.com/leesper/go_rng"
	"github.com/vjeantet/jodaTime"
)

// LooksReal data methods
const (
	IPV4           = "ipv4"
	IPV4CHINA      = "ipv4China"
	CELLPHONECHINA = "cellphoneChina"
	IPV6           = "ipv6"
	MAC            = "mac"
	UA             = "userAgent"
	COUNTRY        = "country"
	EMAIL          = "email"
	NAME           = "name"
	CHINESENAME    = "chineseName"
	UUID           = "uuid"
	XML            = "xml"
	JSON           = "json"
)

// Value selection method
const (
	NEXT   = "next"
	PREV   = "prev"
	RANDOM = "random"
)

// Looks-real opts
const (
	MAXDEPTH    = "maxDepth"
	MAXELEMENTS = "maxElements"
	TAGSEED     = "tagSeed"
)

// seed is the seed data to generate China IP addresses.
var seed = [][]int32{
	{607649792, 608174079},     // 36.56.0.0-36.63.255.255
	{1038614528, 1039007743},   // 61.232.0.0-61.237.255.255
	{1783627776, 1784676351},   // 106.80.0.0-106.95.255.255
	{2035023872, 2035154943},   // 121.76.0.0-121.77.255.255
	{2078801920, 2079064063},   // 123.232.0.0-123.235.255.255
	{-1950089216, -1948778497}, // 139.196.0.0-139.215.255.255
	{-1425539072, -1425014785}, // 171.8.0.0-171.15.255.255
	{-1236271104, -1235419137}, // 182.80.0.0-182.92.255.255
	{-770113536, -768606209},   // 210.25.0.0-210.47.255.255
	{-569376768, -564133889},   // 222.16.0.0-222.95.255.255
}

func init() {
	uuid.Init()
}

// FixedListReplacer is a struct to record config opts of a fixed-list replacement type.
type FixedListReplacer struct {
	method   string
	valRange []string
	currIdx  int
}

// NewFixedListReplacer returns a new FixedListReplacer struct instance
func NewFixedListReplacer(c string, v []string, ci int) Replacer {
	return &FixedListReplacer{
		method:   c,
		valRange: v,
		currIdx:  ci,
	}
}

func (fl *FixedListReplacer) Copy() Replacer {
	n := &FixedListReplacer{
		method:  fl.method,
		currIdx: fl.currIdx,
	}
	n.valRange = append(n.valRange, fl.valRange...)

	return n
}

// ReplacedValue returns a new replacement value of fixed-list type.
func (fl *FixedListReplacer) ReplacedValue(g *rng.GaussianGenerator) (string, error) {
	var newVal string

	switch fl.method {
	case NEXT:
		fl.currIdx = (fl.currIdx + 1) % len(fl.valRange)

	case RANDOM:
		fallthrough
	default:
		fl.currIdx = SimpleGaussian(g, len(fl.valRange))
	}
	newVal = fl.valRange[fl.currIdx]
	return newVal, nil
}

type TimeStampReplacer struct {
	format string
}

// NewTimeStampReplacer returns a new TimeStampReplacer struct instance.
func NewTimeStampReplacer(f string) Replacer {
	return &TimeStampReplacer{
		format: f,
	}
}

func (ts *TimeStampReplacer) Copy() Replacer {
	n := &TimeStampReplacer{
		format: ts.format,
	}

	return n
}

// ReplacedValue populates a new timestamp with current time.
func (ts *TimeStampReplacer) ReplacedValue(*rng.GaussianGenerator) (string, error) {
	return jodaTime.Format(ts.format, time.Now()), nil
}

type StringReplacer struct {
	chars string
	min   int64
	max   int64
}

func NewStringReplacer(chars string, min int64, max int64) Replacer {
	return &StringReplacer{
		chars: chars,
		min:   min,
		max:   max,
	}
}

func (s *StringReplacer) Copy() Replacer {
	n := &StringReplacer{
		chars: s.chars,
		min:   s.min,
		max:   s.max,
	}

	return n
}

func (s *StringReplacer) ReplacedValue(g *rng.GaussianGenerator) (string, error) {
	var str string
	var err error
	if s.min == s.max {
		str = GetRandomString(s.chars, int(s.min))
	} else {
		l := rand.Intn(int(s.max-s.min)) + int(s.min)
		str = GetRandomString(s.chars, l)
	}
	return str, err
}

type FloatReplacer struct {
	min       float64
	max       float64
	precision int64
}

func NewFloatReplacer(min float64, max float64, precision int64) Replacer {
	return &FloatReplacer{
		min:       min,
		max:       max,
		precision: precision,
	}
}

func (f *FloatReplacer) Copy() Replacer {
	n := &FloatReplacer{
		min:       f.min,
		max:       f.max,
		precision: f.precision,
	}
	return n
}

func (f *FloatReplacer) ReplacedValue(g *rng.GaussianGenerator) (string, error) {
	v := f.min + rand.Float64()*(f.max-f.min)
	s := fmt.Sprintf("%%.%df", f.precision)
	return fmt.Sprintf(s, v), nil
}

type IntegerReplacer struct {
	method  string
	min     int64
	max     int64
	currVal int64
}

// NewIntegerReplacer returns a new IntegerReplacer struct instance
func NewIntegerReplacer(c string, minV int64, maxV int64, cv int64) Replacer {
	return &IntegerReplacer{
		method:  c,
		min:     minV,
		max:     maxV,
		currVal: cv,
	}
}

func (i *IntegerReplacer) Copy() Replacer {
	n := &IntegerReplacer{
		method:  i.method,
		min:     i.min,
		max:     i.max,
		currVal: i.currVal,
	}
	return n
}

// ReplacedValue is the main function to populate replacement value of an integer type.
func (i *IntegerReplacer) ReplacedValue(g *rng.GaussianGenerator) (string, error) {
	var currVal = i.currVal

	switch i.method {
	case NEXT:
		i.currVal++
		if i.currVal > i.max {
			i.currVal = i.min
		}
	case PREV:
		i.currVal--
		if i.currVal < i.min {
			i.currVal = i.max
		}
	case RANDOM:
		fallthrough
	default: // Use random by default
		i.currVal = int64(SimpleGaussian(g, int(i.max-i.min))) + i.min
	}
	return strconv.FormatInt(currVal, 10), nil
}

// LooksReal is a struct to record the configured method to generate data.
type LooksReal struct {
	method string
	opts   map[string]interface{} // The opts of a specific looks-real type
}

// NewLooksReal returns a new LooksReal struct instance
func NewLooksReal(m string, p map[string]interface{}) Replacer {
	return &LooksReal{
		method: m,
		opts:   p,
	}
}

func (ia *LooksReal) Copy() Replacer {
	n := &LooksReal{
		method: ia.method,
		// opts should be a bunch of read-only data, so it doesn't get deep-copied here.
		opts: ia.opts,
	}
	return n
}

// ReplacedValue returns random data based on the data type selection.
func (ia *LooksReal) ReplacedValue(g *rng.GaussianGenerator) (data string, err error) {
	switch ia.method {
	case IPV4:
		data = randomdata.IpV4Address()
	case IPV4CHINA:
		data = GetRandomChinaIP(g)
	case IPV6:
		data = randomdata.IpV6Address()
	case UA:
		data = randomdata.UserAgentString()
	case COUNTRY:
		data = randomdata.Country(randomdata.FullCountry)
	case EMAIL:
		data = randomdata.Email()
	case NAME:
		data = randomdata.SillyName()
	case CELLPHONECHINA:
		data = GetRandomChinaCellPhoneNo(g)
	case CHINESENAME:
		data = GetRandomChineseName(g)
	case MAC:
		data = randomdata.MacAddress()
	case UUID:
		data = GetRandomUUID()
	case XML:
		data = RandomXML(ia.opts[MAXDEPTH].(int), ia.opts[MAXELEMENTS].(int), ia.opts[TAGSEED].([]string))
	case JSON:
		data = RandomJSON(ia.opts[MAXDEPTH].(int), ia.opts[MAXELEMENTS].(int), ia.opts[TAGSEED].([]string))
	default:
		err = errors.New(fmt.Sprintf("bad format %s", ia.method))
	}
	return data, err
}

// GetRandomChinaIP returns a random IP address of China.
func GetRandomChinaIP(g *rng.GaussianGenerator) string {
	d1 := SimpleGaussian(g, len(seed))
	d2 := SimpleGaussian(g, int(seed[d1][1]-seed[d1][0]))
	return int2ip(seed[d1][0] + int32(d2))
}

// int2ip is a local helper function to convert the random number to an IP address.
func int2ip(n int32) string {
	var ip = make([]string, 0)
	ip = append(ip, strconv.Itoa(int((n>>24)&0xff)))
	ip = append(ip, strconv.Itoa(int((n>>16)&0xff)))
	ip = append(ip, strconv.Itoa(int((n>>8)&0xff)))
	ip = append(ip, strconv.Itoa(int(n&0xff)))

	return strings.Join(ip, ".")
}

// GetRandomChinaCellPhoneNo returns a random cell phone number starts with 130 - 139
func GetRandomChinaCellPhoneNo(g *rng.GaussianGenerator) string {
	var seed = []string{
		"130", "131", "132", "133", "134", "135", "136", "137", "138", "139",
		"147", "148", "150", "151", "152", "157", " 158", "159", "178", "182",
		"183", "184", "187", "188", "145", "146", "155", "156", "166", "175",
		"176", "185", "186", "141", "149", "153", "173", "174", "177", "180",
		"181"}
	var phone = make([]string, 0)
	phone = append(phone, seed[SimpleGaussian(g, len(seed))])
	phone = append(phone, fmt.Sprintf("%06d", SimpleGaussian(g, 1e8)))
	return strings.Join(phone, "")
}

// GetRandomUUID returns a random UUID V4.
func GetRandomUUID() string {
	return uuid.NewV4().String()
}

// GetRandomChineseName generates random (in gaussian distribution) Chinese name
// picked from the seed.
func GetRandomChineseName(g *rng.GaussianGenerator) string {
	seed := []string{
		"李鸿平", "杨漫宇", "彭聪滨", "王新军", "吴鸣定", "蒋茵果", "李益文", "何子荣", "王志忠", "何联建",
		"邓海国", "李萌雯", "张玉燕", "胡翠芳", "魏亚阳", "黄舒红", "许慧平", "贾文春", "张旭弟", "李汀然",
		"谢姝正", "石紫婷", "张苏辉", "高添生", "张秀昕", "张建杨", "萧倩华", "马佳琦", "钱克然", "张绣英",
		"王东松", "吴娜珉", "孙香曼", "翟文河", "苏长亮", "俞吉秋", "傅海栋", "方波强", "张梦华", "张轩素",
		"孙华姬", "杨萌德", "李芯萍", "胡丽坤", "赵文茂", "谭梦洁", "杨柏磊", "陈荷蕾", "李千平", "李旸芳",
		"孙立锦", "刘晨萍", "孙铮卫", "刘林莲", "唐知玖", "樊向明", "胡厚如", "周斯时", "黄萧根", "张奎平",
		"王雪璐", "谢佳伟", "卜兵泉", "张彭晶", "余楠超", "刘慧美", "叶蕙芳", "高士辉", "黄晶勇", "王荷涛",
		"杨定宇", "徐鑫辉", "林景扬", "张嘉艳", "何昭珊", "陆彩刚", "李洽志", "杨白清", "王衍玲", "蒋林亨",
		"王志青", "夏学华", "钟萧光", "邓慧雅", "刘玉兰", "何薇静", "杨中聪", "王晓英", "李林敏", "胡冲铁",
		"孙炅峰", "江不至", "孟祥海", "姚双东", "李德孟", "龙樱王", "王靓军", "段嵩燕", "申真莹", "李官静",
		"刘迪岩", "王冰姬", "杨三宁", "刘蓉加", "杨胜萃", "陈炜梅", "顾英基", "何尚逸", "冯太春", "高徽员",
		"刘存生", "刘珊艳", "余澜吉", "赵显翁", "李玉硕", "张亦军", "汪晗春", "张原琪", "王晓生", "季志国",
		"吴捷丽", "张东斌", "张韫娟", "李金之", "尹梦波", "张瑞丹", "吴晓宁", "白昌忠", "胡圣凡", "赵泽国",
		"李金飞", "赖一林", "高忠林", "萧敬平", "邓佑兵", "汪晓华", "王红鸿", "张言琪", "许小焓", "郑朝学",
		"李锦武", "徐孙兵", "王白川", "许家峰", "张俏玲", "王采杏", "孙芹杰", "程炳霞", "刘雨彬", "吕晨松",
		"赵金罗", "黄罡卫", "郑纯青", "李弗飞", "黄家斌", "刘聪松", "高海军", "王晓淳", "洪文宽", "马杰晔",
		"章泱辉", "李云梅", "王义龙", "袁亚彬", "袁晓妍", "洪伟燕", "张碧明", "龚惠梅", "杨夏光", "周翀高",
		"张晓庆", "李芝琴", "莫令彦", "周晓运", "林永舒", "徐文耘", "胡永平", "王俊敏", "李苑晓", "高泽平",
		"范豹平", "李云芳", "杨一奎", "李桂平", "白春筠", "冯鹏颖", "陈柔玲", "王云云", "康繁颖", "李余剑",
		"王树淳", "陈大彪", "孙万升", "陈俊凝", "叶昊滨", "唐保燕", "冯家华", "吴玉益", "韩州浩", "安希南",
	}
	return seed[SimpleGaussian(g, len(seed))]
}

// SimpleGaussian returns a random value of Gaussian distribution.
// mean=0.5*the_range, stddev=0.2*the_range
func SimpleGaussian(g *rng.GaussianGenerator, gap int) int {
	if gap == 0 {
		return 0
	}
	return int(math.Abs(g.Gaussian(0.5*float64(gap), 0.2*float64(gap)))) % gap
}

// GetRandomString generates a random string of length n.
func GetRandomString(chars string, length int) string {
	return RandomStr(chars, length)
}

// GetXMLStr returns a randomly generated XML doc in string format
func RandomXML(maxDepth int, maxElements int, seed []string) string {
	doc, err := XMLStr(maxDepth, maxElements, seed)
	if err == nil {
		return doc
	} else {
		return ""
	}
}

// RandomJSON returns a randomly generated JSON doc in string format.
func RandomJSON(maxDepth int, maxElements int, seed []string) string {
	json, err := xj.Convert(strings.NewReader(RandomXML(maxDepth, maxElements, seed)))
	if err != nil {
		return ""
	}
	return json.String()
}

// InitLooksRealParms initiate the parameters map with the default values
func InitLooksRealParms(parms map[string]interface{}, t string) {
	switch t {
	case XML, JSON:
		if _, ok := parms[MAXDEPTH]; !ok {
			parms[MAXDEPTH] = 10
		}
		if _, ok := parms[MAXELEMENTS]; !ok {
			parms[MAXELEMENTS] = 100
		}
		if _, ok := parms[TAGSEED]; !ok {
			parms[TAGSEED] = []string{}
		}
	default:
	}
}
