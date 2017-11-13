package main

import (
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
