package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/kellydunn/golang-geo"
)

func main() {

	// -----
	// ARGUMENTS
	// -----
	var tgt string
	var univ string

	//TODO get lat / long indices for target as arguments
	//TODO get lat / long indices for universe as arguments
	//var verbose bool
	//var nummatches int

	flag.StringVar(&tgt, "target", "", "(Required) Path to target file.")
	flag.StringVar(&univ, "universe", "", "(Required) Path to universe file.")
	//flag.StringVar(&univ, "universe", "", "(Required) Path to universe file.")

	flag.Parse()

	// -----
	// DO STH WITH ALL THOSE NICE ARGUMENTS
	// -----

	////

	fmt.Println("Target file: ", tgt)
	fmt.Println("Universe file: ", univ)

	targetFile, err := os.Open(tgt)
	exitOnError(err, "Oops cannot find target file "+tgt)
	defer targetFile.Close()

	universeFile, err := os.Open(univ)
	exitOnError(err, "Oops cannot find universe file "+univ)
	defer universeFile.Close()

	targetReader := csv.NewReader(targetFile)
	targetReader.Comma = ';' // Use tab-delimited instead of comma
	//targetReader.FieldsPerRecord = -1 // number of expected fields ???
	targetData, err := targetReader.ReadAll()
	exitOnError(err, "Could not read target data")
	targetLatIndex := 6
	targetLonIndex := 7
	//fmt.Println(targetData[0])
	//fmt.Println(targetData[1])

	universeReader := csv.NewReader(universeFile)
	universeReader.Comma = '\t'
	universeData, err := universeReader.ReadAll()
	exitOnError(err, "Could not read universe data")
	universeLatIndex := 3
	universeLonIndex := 2
	//fmt.Println(universeData[0][0], universeData[0][1], universeData[0][2], universeData[0][3])
	//fmt.Println(universeData[1][0], universeData[1][1], universeData[1][2], universeData[1][3])
	//fmt.Println(universeData[2][0], universeData[2][1], universeData[2][2], universeData[2][3])

	//TODO refactor getting point at index to function
	tLatStr := targetData[3][targetLatIndex]
	tLonStr := targetData[3][targetLonIndex]
	tLat, err := strconv.ParseFloat(tLatStr, 64)
	ignoreErrorForNow(err)
	tLon, err := strconv.ParseFloat(tLonStr, 64)
	ignoreErrorForNow(err)

	fmt.Println("Target coordinates: ", tLatStr, ",", tLonStr)
	p := geo.NewPoint(tLat, tLon)

	//nearest := universeData[1]
	var nearest []string
	var lowestDistance float64
	//lowestDistance = -1.0
	var maxUIndex int
	if universeLatIndex > universeLonIndex {
		maxUIndex = universeLatIndex
	} else {
		maxUIndex = universeLonIndex
	}
	for key, record := range universeData {
		// skip headers
		if key < 1 {
			continue
		}
		l := len(record)
		if l <= maxUIndex {
			fmt.Println("Problem getting uLatLon at line ", key+1, "")
			continue
		}
		uLatStr := record[universeLatIndex]
		uLonStr := record[universeLonIndex]
		uLat, err := strconv.ParseFloat(uLatStr, 64)
		ignoreErrorForNow(err)
		uLon, err := strconv.ParseFloat(uLonStr, 64)
		ignoreErrorForNow(err)
		up := geo.NewPoint(uLat, uLon)
		dist := p.GreatCircleDistance(up)
		if key == 1 {
			lowestDistance = dist
			nearest = record
			fmt.Println("Init lowestDistance to ", lowestDistance)
			continue
		}
		if dist < lowestDistance {
			lowestDistance = dist
			nearest = record
		}
	}

	header := append(targetData[0], universeData[0]...)
	result := append(targetData[3], nearest...)
	var results [][]string
	results = append(results, header)
	results = append(results, result)

	fmt.Println("Nearest and dearest")
	fmt.Println(nearest)
	fmt.Println(lowestDistance)
	fmt.Println(header)
	fmt.Println(result)
	fmt.Println(results)

	/*
		for _, record := range targetData {
			fmt.Println(record)
		}
	*/
	//In paulmach/go.geo ...
	//p1 := geo.NewPointFromLatLng(51.92125, 6.57755)
	//p2 := geo.NewPointFromLatLng(52.377777, 4.905169)
	//d := p1.GeoDistanceFrom(p2)

}

// exitOnError exits with an error message if error is not nil
func exitOnError(e error, msg string) {
	if e != nil {
		fmt.Println(msg)
		os.Exit(1)
	}
}

func ignoreErrorForNow(e error) {
	return
}

func checkStuff() {

	// TODO move checks to separate file
	//In kellydunn/golang-geo a point is LAT / LONG
	p1 := geo.NewPoint(51.92125, 6.57755)
	p2 := geo.NewPoint(52.377777, 4.905169)

	d := p1.GreatCircleDistance(p2)
	fmt.Println("Test distance 1: ", d)
	p3a := geo.NewPoint(51.92126, 6.57756)
	p3b := geo.NewPoint(51.921598, 6.577753)
	d = p3a.GreatCircleDistance(p3b)
	fmt.Println("Test distance 2: ", d)

	p4x, err := strconv.ParseFloat("51.92125", 64)
	ignoreErrorForNow(err)
	p4y, err := strconv.ParseFloat("6.57755", 64)
	ignoreErrorForNow(err)
	p4 := geo.NewPoint(p4x, p4y)

	p5x, err := strconv.ParseFloat("52.377777", 64)
	ignoreErrorForNow(err)
	p5y, err := strconv.ParseFloat("4.905169", 64)
	ignoreErrorForNow(err)
	p5 := geo.NewPoint(p5x, p5y)
	d = p4.GreatCircleDistance(p5)
	fmt.Println("Test distance 3: ", d)

}
