package main

import (
	"fmt"
	"github.com/leesper/go_rng"
	"strconv"
	"strings"
)

// seed is the seed data to generate China IP addresses.
var seed = [][]int32{
	{607649792, 608174079},     //36.56.0.0-36.63.255.255
	{1038614528, 1039007743},   //61.232.0.0-61.237.255.255
	{1783627776, 1784676351},   //106.80.0.0-106.95.255.255
	{2035023872, 2035154943},   //121.76.0.0-121.77.255.255
	{2078801920, 2079064063},   //123.232.0.0-123.235.255.255
	{-1950089216, -1948778497}, //139.196.0.0-139.215.255.255
	{-1425539072, -1425014785}, //171.8.0.0-171.15.255.255
	{-1236271104, -1235419137}, //182.80.0.0-182.92.255.255
	{-770113536, -768606209},   //210.25.0.0-210.47.255.255
	{-569376768, -564133889},   //222.16.0.0-222.95.255.255
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
		"181", "199"}
	var phone = make([]string, 0)
	phone = append(phone, seed[SimpleGaussian(g, len(seed))])
	phone = append(phone, fmt.Sprintf("%06d", SimpleGaussian(g, 1e8)))
	return strings.Join(phone, "")
}

// GetRandomChineseName generates random (in gaussian distribution) Chinese name
// picked from the seed.
func GetRandomChineseName(g *rng.GaussianGenerator) string {
	seed := `
	李鸿平, 杨漫宇, 彭聪滨, 王新军, 吴鸣定, 蒋茵果, 李益文, 何子荣, 王志忠, 何联建,
	邓海国, 李萌雯, 张玉燕, 胡翠芳, 魏亚阳, 黄舒红, 许慧平, 贾文春, 张旭弟, 李汀然,
	谢姝正, 石紫婷, 张苏辉, 高添生, 张秀昕, 张建杨, 萧倩华, 马佳琦, 钱克然, 张绣英,
	王东松, 吴娜珉, 孙香曼, 翟文河, 苏长亮, 俞吉秋, 傅海栋, 方波强, 张梦华, 张轩素,
	孙华姬, 杨萌德, 李芯萍, 胡丽坤, 赵文茂, 谭梦洁, 杨柏磊, 陈荷蕾, 李千平, 李旸芳,
	孙立锦, 刘晨萍, 孙铮卫, 刘林莲, 唐知玖, 樊向明, 胡厚如, 周斯时, 黄萧根, 张奎平,
	王雪璐, 谢佳伟, 卜兵泉, 张彭晶, 余楠超, 刘慧美, 叶蕙芳, 高士辉, 黄晶勇, 王荷涛,
	杨定宇, 徐鑫辉, 林景扬, 张嘉艳, 何昭珊, 陆彩刚, 李洽志, 杨白清, 王衍玲, 蒋林亨,
	王志青, 夏学华, 钟萧光, 邓慧雅, 刘玉兰, 何薇静, 杨中聪, 王晓英, 李林敏, 胡冲铁,
	孙炅峰, 江不至, 孟祥海, 姚双东, 李德孟, 龙樱王, 王靓军, 段嵩燕, 申真莹, 李官静,
	刘迪岩, 王冰姬, 杨三宁, 刘蓉加, 杨胜萃, 陈炜梅, 顾英基, 何尚逸, 冯太春, 高徽员,
	刘存生, 刘珊艳, 余澜吉, 赵显翁, 李玉硕, 张亦军, 汪晗春, 张原琪, 王晓生, 季志国,
	吴捷丽, 张东斌, 张韫娟, 李金之, 尹梦波, 张瑞丹, 吴晓宁, 白昌忠, 胡圣凡, 赵泽国,
	李金飞, 赖一林, 高忠林, 萧敬平, 邓佑兵, 汪晓华, 王红鸿, 张言琪, 许小焓, 郑朝学,
	李锦武, 徐孙兵, 王白川, 许家峰, 张俏玲, 王采杏, 孙芹杰, 程炳霞, 刘雨彬, 吕晨松,
	赵金罗, 黄罡卫, 郑纯青, 李弗飞, 黄家斌, 刘聪松, 高海军, 王晓淳, 洪文宽, 马杰晔,
	章泱辉, 李云梅, 王义龙, 袁亚彬, 袁晓妍, 洪伟燕, 张碧明, 龚惠梅, 杨夏光, 周翀高,
	张晓庆, 李芝琴, 莫令彦, 周晓运, 林永舒, 徐文耘, 胡永平, 王俊敏, 李苑晓, 高泽平,
	范豹平, 李云芳, 杨一奎, 李桂平, 白春筠, 冯鹏颖, 陈柔玲, 王云云, 康繁颖, 李余剑,
	王树淳, 陈大彪, 孙万升, 陈俊凝, 叶昊滨, 唐保燕, 冯家华, 吴玉益, 韩州浩, 安希南
	`
	names := strings.Split(seed, ",")
	for i := 0; i < len(names); i++ {
		names[i] = strings.Trim(names[i], " \n\t")
	}
	return names[SimpleGaussian(g, len(names))]
}
