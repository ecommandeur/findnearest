package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/paulmach/go.geo"
)

func main() {

	// -----
	// ARGUMENTS
	// -----
	var tgt string
	var univ string

	//TODO get lat / lngg indices for target as arguments
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
	fmt.Println("Output file: result.csv")

	targetFile, err := os.Open(tgt)
	exitOnError(err, "Oops, cannot find target file "+tgt)
	defer targetFile.Close()

	universeFile, err := os.Open(univ)
	exitOnError(err, "Oops, cannot find universe file "+univ)
	defer universeFile.Close()

	outputFile, err := os.Create("result.csv")
	exitOnError(err, "Oops, cannot create result file.")
	defer outputFile.Close()

	targetReader := csv.NewReader(targetFile)
	targetReader.Comma = ';' // Use semi-colon instead of comma
	//targetReader.FieldsPerRecord = -1 // number of expected fields ???
	targetData, err := targetReader.ReadAll()
	exitOnError(err, "Could not read target data")
	targetLatIndex := 6
	targetLngIndex := 7

	universeReader := csv.NewReader(universeFile)
	universeReader.Comma = '\t' // Use tab instead of comma
	universeData, err := universeReader.ReadAll()
	exitOnError(err, "Could not read universe data")
	universeLatIndex := 3
	universeLngIndex := 2

	var results [][]string
	var nearest []string
	var lowestDistance float64
	maxTIndex := max(targetLatIndex, targetLngIndex)
	maxUIndex := max(universeLatIndex, universeLngIndex)
	header := append(targetData[0], universeData[0]...) //TODO allow files without header
	results = append(results, header)
	for k, r := range targetData {
		//TODO allow files without header
		if k < 1 {
			continue
		}
		l := len(r)
		if l <= maxTIndex {
			fmt.Println("Target LatLng index out of range at line ", k+1, "")
			continue
		}
		p, err := newPointFromLatLngStrings(r[targetLatIndex], r[targetLngIndex])
		if err != nil {
			fmt.Println(err.Error(), " at line ", k+1)
			continue
		}
		//TODO see if we can do this in a smarter way. For now, do it brute force
		// for each target record, look at all universeData
		for key, record := range universeData {
			// skip headers
			if key < 1 {
				continue
			}
			l := len(record)
			if l <= maxUIndex {
				fmt.Println("Universe LatLng index out of range at line ", key+1, "")
				continue
			}
			up, err := newPointFromLatLngStrings(record[universeLatIndex], record[universeLngIndex])
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			dist := p.GeoDistanceFrom(up, true)
			if key == 1 {
				lowestDistance = dist
				nearest = record
				//fmt.Println("Init lowestDistance to ", lowestDistance)
				continue
			}
			if dist < lowestDistance {
				lowestDistance = dist
				nearest = record
			}
		}
		result := append(r, nearest...)
		results = append(results, result)
	}

	//fmt.Println("Nearest and dearest")
	//fmt.Println(results)

	csvWriter := csv.NewWriter(outputFile)
	csvWriter.WriteAll(results)
}

// newPointFromLatLngStrings takes two strings that are supposedly lat/lng
// and tries to create a Point from those strings
func newPointFromLatLngStrings(latStr string, lngStr string) (*geo.Point, error) {
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	var p *geo.Point
	if err1 != nil || err2 != nil {
		msg := []string{"Unable to convert ", latStr, ",", lngStr, " to a Point"}
		return nil, errors.New(strings.Join(msg, ""))
	}
	p = geo.NewPointFromLatLng(lat, lng)
	return p, nil
}

// exitOnError exits with an error message if error is not nil
func exitOnError(e error, msg string) {
	if e != nil {
		fmt.Println(msg)
		os.Exit(1)
	}
}

// get the maximum of two integers
func max(x int, y int) int {
	if x > y {
		return x
	}
	return y
}
