package utils

import "math"

// GeoDistance
// 球面距离公式：https://baike.baidu.com/item/%E7%90%83%E9%9D%A2%E8%B7%9D%E7%A6%BB%E5%85%AC%E5%BC%8F/5374455?fr=aladdin
// GeoDistance 计算地理距离，依次为两个坐标的纬度、经度, 单位:公里
func GeoDistance(lng1 float64, lat1 float64, lng2 float64, lat2 float64) float64 {
	if lng1 == 0 || lat1 == 0 || lng2 == 0 || lat2 == 0 {
		return -0.1
	}

	const PI float64 = 3.141592653589793

	radlat1 := float64(PI * lat1 / 180)
	radlat2 := float64(PI * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(PI * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515
	return dist * 1.609344
}
