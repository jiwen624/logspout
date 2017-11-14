package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

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
func GetRandomChinaIP() string {
	d1 := rand.Int31n(int32(len(seed)))
	d2 := rand.Int31n(seed[d1][1] - seed[d1][0])
	return int2ip(seed[d1][0] + d2)
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
func GetRandomChinaCellPhoneNo() string {
	var seed = []string{
		"130", "131", "132", "133", "134", "135", "136", "137", "138", "139",
		"147", "148", "150", "151", "152", "157", " 158", "159", "178", "182",
		"183", "184", "187", "188", "145", "146", "155", "156", "166", "175",
		"176", "185", "186", "141", "149", "153", "173", "174", "177", "180",
		"181", "199"}
	var phone = make([]string, 0)
	phone = append(phone, seed[rand.Intn(len(seed))])
	phone = append(phone, fmt.Sprintf("%06d", rand.Intn(100000000)))
	return strings.Join(phone, "")
}
