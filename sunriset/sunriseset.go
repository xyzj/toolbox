// Thanks https://github.com/kelvins/sunrisesunset
//
// Package sunrisesunset should be used to calculate the apparent sunrise and sunset based on the latitude, longitude, UTC offset and date.
// All calculations (formulas) were extracted from the Solar Calculation Details of the Earth System Research Laboratory:
// https://www.esrl.noaa.gov/gmd/grad/solcalc/calcdetails.html

// Package sunriset 日出日落功能模块
package sunriset

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xyzj/toolbox/mapfx"
)

// Result 日出日落结果
type Result struct {
	Month   int
	Day     int
	Sunrise int
	Sunset  int
}

// Params The Parameters struct can also be used to manipulate the data and get the sunrise and sunset
// sunrise/sunset int hh*60+mm
type Params struct {
	sunResult *mapfx.StructMap[string, Result]
	Latitude  float64
	Longitude float64
	UtcOffset float64
	Year      int
}

// ForEach 遍历计算结果集
func (p *Params) ForEach(f func(month, day, sunrise, sunset int) bool) {
	p.sunResult.ForEach(func(key string, value *Result) bool {
		return f(value.Month, value.Day, value.Sunrise, value.Sunset)
	})
}

// Calculation 使用当前年份计算全年日出日落时间hh*60+mm，非润年，2月29日采用28日时间
func (p *Params) Calculation() bool {
	if p.sunResult == nil {
		p.sunResult = mapfx.NewStructMap[string, Result]()
	}
	if p.Year == 0 {
		p.Year = time.Now().UTC().Year()
	}
	if p.UtcOffset == 0 {
		_, tz := time.Now().Zone()
		p.UtcOffset = float64(tz) / 3600
	}
	var lock sync.WaitGroup
	var faultCount int32
	var days int32
	for i := 1; i <= 366; i++ {
		lock.Add(1)
		go func() {
			defer lock.Done()
			dd := int(atomic.AddInt32(&days, 1))
			tt := time.Date(p.Year, time.January, dd, 0, 0, 0, 0, time.Local)
			rise, set, err := GetSunriseSunset(p.Latitude, p.Longitude, p.UtcOffset, tt)
			if err != nil {
				atomic.AddInt32(&faultCount, 1)
				return
			}
			p.sunResult.Store(fmt.Sprintf("%02d%02d", tt.Month(), tt.Day()), &Result{
				Month:   int(tt.Month()),
				Day:     tt.Day(),
				Sunrise: rise.Hour()*60 + rise.Minute(),
				Sunset:  set.Hour()*60 + set.Minute(),
			})
		}()
	}
	lock.Wait()
	if faultCount > 1 {
		return false
	}
	feb, ok := p.sunResult.Load("0228")
	if ok {
		feb29 := &Result{
			Month:   2,
			Day:     29,
			Sunrise: feb.Sunrise,
			Sunset:  feb.Sunset,
		}
		p.sunResult.Store("0229", feb29)
	} else {
		return false
	}
	return true
}

// Get 获取指定月日的日出日落时间
func (p *Params) Get(month, day int) (rise, set int) {
	if month > 12 || month < 1 || day < 1 || day > 31 {
		return 0, 0
	}
	r, ok := p.sunResult.Load(fmt.Sprintf("%02d%02d", month, day))
	if ok {
		return r.Sunrise, r.Sunset
	}
	return 0, 0
}

// Convert radians to degrees
func rad2deg(radians float64) float64 {
	return radians * (180.0 / math.Pi)
}

// Convert degrees to radians
func deg2rad(degrees float64) float64 {
	return degrees * (math.Pi / 180.0)
}

// Creates a vector with the seconds normalized to the range 0~1.
// seconds - The number of seconds will be normalized to 1
// Return A vector with the seconds normalized to 0~1
func createSecondsNormalized(seconds int) (vector []float64) {
	vector = make([]float64, seconds)
	for index := 0; index < seconds; index++ {
		vector[index] = float64(index) / float64(seconds-1)
	}
	return
}

// Calculate Julian Day based on the formula: nDays+2415018.5+secondsNorm-UTCoff/24
// numDays - The number of days calculated in the calculate function
// secondsNorm - Seconds normalized calculated by the createSecondsNormalized function
// utcOffset - UTC offset defined by the user
// Return Julian day slice
func calcJulianDay(numDays int64, secondsNorm []float64, utcOffset float64) (julianDay []float64) {
	julianDay = make([]float64, len(secondsNorm))
	for index := 0; index < len(secondsNorm); index++ {
		julianDay[index] = float64(numDays) + 2415018.5 + secondsNorm[index] - utcOffset/24.0
	}
	return
}

// Calculate the Julian Century based on the formula: (julianDay - 2451545.0) / 36525.0
// julianDay - Julian day vector calculated by the calcJulianDay function
// Return Julian century slice
func calcJulianCentury(julianDay []float64) (julianCentury []float64) {
	julianCentury = make([]float64, len(julianDay))
	for index := 0; index < len(julianDay); index++ {
		julianCentury[index] = (julianDay[index] - 2451545.0) / 36525.0
	}
	return
}

// Calculate the Geom Mean Long Sun in degrees based on the formula: 280.46646 + julianCentury * (36000.76983 + julianCentury * 0.0003032)
// julianCentury - Julian century calculated by the calcJulianCentury function
// Return The Geom Mean Long Sun slice
func calcGeomMeanLongSun(julianCentury []float64) (geomMeanLongSun []float64) {
	geomMeanLongSun = make([]float64, len(julianCentury))
	for index := 0; index < len(julianCentury); index++ {
		a := 280.46646 + julianCentury[index]*(36000.76983+julianCentury[index]*0.0003032)
		geomMeanLongSun[index] = math.Mod(a, 360.0)
	}
	return
}

// Calculate the Geom Mean Anom Sun in degrees based on the formula: 357.52911 + julianCentury * (35999.05029 - 0.0001537 * julianCentury)
// julianCentury - Julian century calculated by the calcJulianCentury function
// Return The Geom Mean Anom Sun slice
func calcGeomMeanAnomSun(julianCentury []float64) (geomMeanAnomSun []float64) {
	geomMeanAnomSun = make([]float64, len(julianCentury))
	for index := 0; index < len(julianCentury); index++ {
		geomMeanAnomSun[index] = 357.52911 + julianCentury[index]*(35999.05029-0.0001537*julianCentury[index])
	}
	return
}

// Calculate the Eccent Earth Orbit based on the formula: 0.016708634 - julianCentury * (0.000042037 + 0.0000001267 * julianCentury)
// julianCentury - Julian century calculated by the calcJulianCentury function
// Return The Eccent Earth Orbit slice
func calcEccentEarthOrbit(julianCentury []float64) (eccentEarthOrbit []float64) {
	eccentEarthOrbit = make([]float64, len(julianCentury))
	for index := 0; index < len(julianCentury); index++ {
		eccentEarthOrbit[index] = 0.016708634 - julianCentury[index]*(0.000042037+0.0000001267*julianCentury[index])
	}
	return
}

// Calculate the Sun Eq Ctr based on the formula: sin(deg2rad(geomMeanAnomSun))*(1.914602-julianCentury*(0.004817+0.000014*julianCentury))+sin(deg2rad(2*geomMeanAnomSun))*(0.019993-0.000101*julianCentury)+sin(deg2rad(3*geomMeanAnomSun))*0.000289;
// julianCentury - Julian century calculated by the calcJulianCentury function
// geomMeanAnomSun - Geom Mean Anom Sun calculated by the calcGeomMeanAnomSun function
// Return The Sun Eq Ctr slice
func calcSunEqCtr(julianCentury []float64, geomMeanAnomSun []float64) (sunEqCtr []float64) {
	if len(julianCentury) != len(geomMeanAnomSun) {
		return
	}

	sunEqCtr = make([]float64, len(julianCentury))
	for index := 0; index < len(julianCentury); index++ {
		sunEqCtr[index] = math.Sin(deg2rad(geomMeanAnomSun[index]))*(1.914602-julianCentury[index]*(0.004817+0.000014*julianCentury[index])) + math.Sin(deg2rad(2*geomMeanAnomSun[index]))*(0.019993-0.000101*julianCentury[index]) + math.Sin(deg2rad(3*geomMeanAnomSun[index]))*0.000289
	}
	return
}

// Calculate the Sun True Long in degrees based on the formula: sunEqCtr + geomMeanLongSun
// sunEqCtr - Sun Eq Ctr calculated by the calcSunEqCtr function
// geomMeanLongSun - Geom Mean Long Sun calculated by the calcGeomMeanLongSun function
// Return The Sun True Long slice
func calcSunTrueLong(sunEqCtr []float64, geomMeanLongSun []float64) (sunTrueLong []float64) {
	if len(sunEqCtr) != len(geomMeanLongSun) {
		return
	}

	sunTrueLong = make([]float64, len(sunEqCtr))
	for index := 0; index < len(sunEqCtr); index++ {
		sunTrueLong[index] = sunEqCtr[index] + geomMeanLongSun[index]
	}
	return
}

// Calculate the Sun App Long in degrees based on the formula: sunTrueLong-0.00569-0.00478*sin(deg2rad(125.04-1934.136*julianCentury))
// sunTrueLong - Sun True Long calculated by the calcSunTrueLong function
// julianCentury - Julian century calculated by the calcJulianCentury function
// Return The Sun App Long slice
func calcSunAppLong(sunTrueLong []float64, julianCentury []float64) (sunAppLong []float64) {
	if len(sunTrueLong) != len(julianCentury) {
		return
	}

	sunAppLong = make([]float64, len(sunTrueLong))
	for index := 0; index < len(sunTrueLong); index++ {
		sunAppLong[index] = sunTrueLong[index] - 0.00569 - 0.00478*math.Sin(deg2rad(125.04-1934.136*julianCentury[index]))
	}
	return
}

// Calculate the Mean Obliq Ecliptic in degrees based on the formula: 23+(26+((21.448-julianCentury*(46.815+julianCentury*(0.00059-julianCentury*0.001813))))/60)/60
// julianCentury - Julian century calculated by the calcJulianCentury function
// Return the Mean Obliq Ecliptic slice
func calcMeanObliqEcliptic(julianCentury []float64) (meanObliqEcliptic []float64) {
	meanObliqEcliptic = make([]float64, len(julianCentury))
	for index := 0; index < len(julianCentury); index++ {
		meanObliqEcliptic[index] = 23.0 + (26.0+(21.448-julianCentury[index]*(46.815+julianCentury[index]*(0.00059-julianCentury[index]*0.001813)))/60.0)/60.0
	}
	return
}

// Calculate the Obliq Corr in degrees based on the formula: meanObliqEcliptic+0.00256*cos(deg2rad(125.04-1934.136*julianCentury))
// meanObliqEcliptic - Mean Obliq Ecliptic calculated by the calcMeanObliqEcliptic function
// julianCentury - Julian century calculated by the calcJulianCentury function
// Return the Obliq Corr slice
func calcObliqCorr(meanObliqEcliptic []float64, julianCentury []float64) (obliqCorr []float64) {
	if len(meanObliqEcliptic) != len(julianCentury) {
		return
	}

	obliqCorr = make([]float64, len(julianCentury))
	for index := 0; index < len(julianCentury); index++ {
		obliqCorr[index] = meanObliqEcliptic[index] + 0.00256*math.Cos(deg2rad(125.04-1934.136*julianCentury[index]))
	}
	return
}

// Calculate the Sun Declination in degrees based on the formula: rad2deg(asin(sin(deg2rad(obliqCorr))*sin(deg2rad(sunAppLong))))
// obliqCorr - Obliq Corr calculated by the calcObliqCorr function
// sunAppLong - Sun App Long calculated by the calcSunAppLong function
// Return the sun declination slice
func calcSunDeclination(obliqCorr []float64, sunAppLong []float64) (sunDeclination []float64) {
	if len(obliqCorr) != len(sunAppLong) {
		return
	}

	sunDeclination = make([]float64, len(obliqCorr))
	for index := 0; index < len(obliqCorr); index++ {
		sunDeclination[index] = rad2deg(math.Asin(math.Sin(deg2rad(obliqCorr[index])) * math.Sin(deg2rad(sunAppLong[index]))))
	}
	return
}

// Calculate the equation of time (minutes) based on the formula:
// 4*rad2deg(multiFactor*sin(2*deg2rad(geomMeanLongSun))-2*eccentEarthOrbit*sin(deg2rad(geomMeanAnomSun))+4*eccentEarthOrbit*multiFactor*sin(deg2rad(geomMeanAnomSun))*cos(2*deg2rad(geomMeanLongSun))-0.5*multiFactor*multiFactor*sin(4*deg2rad(geomMeanLongSun))-1.25*eccentEarthOrbit*eccentEarthOrbit*sin(2*deg2rad(geomMeanAnomSun)))
// multiFactor - The Multi Factor vector calculated in the calculate function
// geomMeanLongSun - The Geom Mean Long Sun vector calculated by the calcGeomMeanLongSun function
// eccentEarthOrbit - The Eccent Earth vector calculated by the calcEccentEarthOrbit function
// geomMeanAnomSun - The Geom Mean Anom Sun vector calculated by the calcGeomMeanAnomSun function
// Return the equation of time slice
func calcEquationOfTime(multiFactor []float64, geomMeanLongSun []float64, eccentEarthOrbit []float64, geomMeanAnomSun []float64) (equationOfTime []float64) {
	if len(multiFactor) != len(geomMeanLongSun) ||
		len(multiFactor) != len(eccentEarthOrbit) ||
		len(multiFactor) != len(geomMeanAnomSun) {
		return
	}

	equationOfTime = make([]float64, len(multiFactor))
	for index := 0; index < len(multiFactor); index++ {
		a := multiFactor[index] * math.Sin(2.0*deg2rad(geomMeanLongSun[index]))
		b := 2.0 * eccentEarthOrbit[index] * math.Sin(deg2rad(geomMeanAnomSun[index]))
		c := 4.0 * eccentEarthOrbit[index] * multiFactor[index] * math.Sin(deg2rad(geomMeanAnomSun[index]))
		d := math.Cos(2.0 * deg2rad(geomMeanLongSun[index]))
		e := 0.5 * multiFactor[index] * multiFactor[index] * math.Sin(4.0*deg2rad(geomMeanLongSun[index]))
		f := 1.25 * eccentEarthOrbit[index] * eccentEarthOrbit[index] * math.Sin(2.0*deg2rad(geomMeanAnomSun[index]))
		equationOfTime[index] = 4.0 * rad2deg(a-b+c*d-e-f)
	}
	return
}

// Calculate the HaSunrise in degrees based on the formula: rad2deg(acos(cos(deg2rad(90.833))/(cos(deg2rad(latitude))*cos(deg2rad(sunDeclination)))-tan(deg2rad(latitude))*tan(deg2rad(sunDeclination))))
// latitude - The latitude defined by the user
// sunDeclination - The Sun Declination calculated by the calcSunDeclination function
// Return the HaSunrise slice
func calcHaSunrise(latitude float64, sunDeclination []float64) (haSunrise []float64) {
	haSunrise = make([]float64, len(sunDeclination))
	for index := 0; index < len(sunDeclination); index++ {
		haSunrise[index] = rad2deg(math.Acos(math.Cos(deg2rad(90.833))/(math.Cos(deg2rad(latitude))*math.Cos(deg2rad(sunDeclination[index]))) - math.Tan(deg2rad(latitude))*math.Tan(deg2rad(sunDeclination[index]))))
	}
	return
}

// Calculate the Solar Noon based on the formula: (720 - 4 * longitude - equationOfTime + utcOffset * 60) * 60
// longitude - The longitude is defined by the user
// equationOfTime - The Equation of Time slice is calculated by the calcEquationOfTime function
// utcOffset - The UTC offset is defined by the user
// Return the Solar Noon slice
func calcSolarNoon(longitude float64, equationOfTime []float64, utcOffset float64) (solarNoon []float64) {
	solarNoon = make([]float64, len(equationOfTime))
	for index := 0; index < len(equationOfTime); index++ {
		solarNoon[index] = (720.0 - 4.0*longitude - equationOfTime[index] + utcOffset*60.0) * 60.0
	}
	return
}

// Check if the latitude is valid. Range: -90 - 90
func checkLatitude(latitude float64) bool {
	if latitude < -90.0 || latitude > 90.0 {
		return false
	}
	return true
}

// Check if the longitude is valid. Range: -180 - 180
func checkLongitude(longitude float64) bool {
	if longitude < -180.0 || longitude > 180.0 {
		return false
	}
	return true
}

// Check if the UTC offset is valid. Range: -12 - 14
func checkUtcOffset(utcOffset float64) bool {
	if utcOffset < -12.0 || utcOffset > 14.0 {
		return false
	}
	return true
}

// Check if the date is valid.
func checkDate(date time.Time) bool {
	minDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	maxDate := time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC)
	if date.Before(minDate) || date.After(maxDate) {
		return false
	}
	return true
}

// Compute the number of days between two dates
func diffDays(date1, date2 time.Time) int64 {
	return int64(date2.Sub(date1) / (24 * time.Hour))
}

// Find the index of the minimum value
func minIndex(slice []float64) int {
	if len(slice) == 0 {
		return -1
	}
	min := slice[0]
	minIndex := 0
	for index := 0; index < len(slice); index++ {
		if slice[index] < min {
			min = slice[index]
			minIndex = index
		}
	}
	return minIndex
}

// Convert each value to the absolute value
func abs(slice []float64) []float64 {
	newSlice := make([]float64, len(slice))
	index := 0
	for _, value := range slice {
		if value < 0.0 {
			value = math.Abs(value)
		}
		newSlice[index] = value
		index++
	}
	return newSlice
}

func round(value float64) int {
	if value < 0 {
		return int(value - 0.5)
	}
	return int(value + 0.5)
}

// GetSunriseSunset function is responsible for calculate the apparent Sunrise and Sunset times.
// If some parameter is wrong it will return an error.
func GetSunriseSunset(latitude float64, longitude float64, utcOffset float64, date time.Time) (sunrise time.Time, sunset time.Time, err error) {
	// Check latitude
	if !checkLatitude(latitude) {
		err = errors.New("invalid latitude")
		return
	}
	// Check longitude
	if !checkLongitude(longitude) {
		err = errors.New("invalid longitude")
		return
	}
	// Check UTC offset
	if !checkUtcOffset(utcOffset) {
		err = errors.New("invalid UTC offset")
		return
	}
	// Check date
	if !checkDate(date) {
		err = errors.New("invalid date")
		return
	}

	// The number of days since 30/12/1899
	since := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	numDays := diffDays(since, date)

	// Seconds of a full day 86400
	seconds := 24 * 60 * 60

	// Creates a vector that represents each second in the range 0~1
	secondsNorm := createSecondsNormalized(seconds)

	// Calculate Julian Day
	julianDay := calcJulianDay(numDays, secondsNorm, utcOffset)

	// Calculate Julian Century
	julianCentury := calcJulianCentury(julianDay)

	// Geom Mean Long Sun (deg)
	geomMeanLongSun := calcGeomMeanLongSun(julianCentury)

	// Geom Mean Anom Sun (deg)
	geomMeanAnomSun := calcGeomMeanAnomSun(julianCentury)

	// Eccent Earth Orbit
	eccentEarthOrbit := calcEccentEarthOrbit(julianCentury)

	// Sun Eq of Ctr
	sunEqCtr := calcSunEqCtr(julianCentury, geomMeanAnomSun)

	// Sun True Long (deg)
	sunTrueLong := calcSunTrueLong(sunEqCtr, geomMeanLongSun)

	// Sun App Long (deg)
	sunAppLong := calcSunAppLong(sunTrueLong, julianCentury)

	// Mean Obliq Ecliptic (deg)
	meanObliqEcliptic := calcMeanObliqEcliptic(julianCentury)

	// Obliq Corr (deg)
	obliqCorr := calcObliqCorr(meanObliqEcliptic, julianCentury)

	// Sun Declin (deg)
	sunDeclination := calcSunDeclination(obliqCorr, sunAppLong)

	// var y
	multiFactor := make([]float64, len(obliqCorr))
	for index := 0; index < len(obliqCorr); index++ {
		temp := math.Tan(deg2rad(obliqCorr[index]/2.0)) * math.Tan(deg2rad(obliqCorr[index]/2.0))
		multiFactor[index] = temp
	}

	// Eq of Time (minutes)
	equationOfTime := calcEquationOfTime(multiFactor, geomMeanLongSun, eccentEarthOrbit, geomMeanAnomSun)

	// HA Sunrise (deg)
	haSunrise := calcHaSunrise(latitude, sunDeclination)

	// Solar Noon (LST)
	solarNoon := calcSolarNoon(longitude, equationOfTime, utcOffset)

	// Sunrise and Sunset Times (LST)
	tempSunrise := make([]float64, len(solarNoon))
	tempSunset := make([]float64, len(solarNoon))

	for index := 0; index < len(solarNoon); index++ {
		tempSunrise[index] = (solarNoon[index] - float64(round(haSunrise[index]*4.0*60.0)) - float64(seconds)*secondsNorm[index])
		tempSunset[index] = (solarNoon[index] + float64(round(haSunrise[index]*4.0*60.0)) - float64(seconds)*secondsNorm[index])
	}

	// Get the sunrise and sunset in seconds
	sunriseSeconds := minIndex(abs(tempSunrise))
	sunsetSeconds := minIndex(abs(tempSunset))

	// Convert the seconds to time
	defaultTime := new(time.Time)
	sunrise = defaultTime.Add(time.Duration(sunriseSeconds) * time.Second)
	sunset = defaultTime.Add(time.Duration(sunsetSeconds+60) * time.Second) // 手动修正1分钟

	return
}
