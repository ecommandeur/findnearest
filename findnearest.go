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

//TODO move initialization of flags to init function
// see https://github.com/spiffytech/csvmaster/blob/master/csvmaster.go

//TODO rewrite to leverage spatial index

func main() {

	const VERSION = "0.2-SNAPSHOT"

	// -----
	// ARGUMENTS
	// -----
	var tgt string
	var univ string
	var tlat, tlng int
	var ulat, ulng int
	var tsep, usep string
	var printVersion bool
	//var printHelp bool

	//var verbose bool
	//TODO support more than one match
	//var nummatches int

	flag.StringVar(&tgt, "target", "", "(Required) Path to target file.")
	flag.StringVar(&univ, "universe", "", "(Required) Path to universe file.")
	flag.IntVar(&tlat, "tlat", 0, "(Required) Index of Latitude column in target file.")
	flag.IntVar(&tlng, "tlng", 0, "(Required) Index of Latitude column in target file.")

	//flag.StringVar(&tSep, "tsep", "(Required) Field separator in target file.")
	flag.IntVar(&ulat, "ulat", 0, "(Required) Index of Latitude column in universe file.")
	flag.IntVar(&ulng, "ulng", 0, "(Required) Index of Longitude column in universe file.")

	flag.StringVar(&tsep, "tsep", ",", "(Optional) Field separator in target file ('tab' for tab-separated).")
	flag.StringVar(&usep, "usep", ",", "(Optional) Field separator in universe file ('tab' for tab-separated).")

	flag.BoolVar(&printVersion, "version", false, "Print program version")
	//flag.BoolVar(&printHelp, "h", false, "Print help")

	flag.Parse()

	// -----
	// DO STH WITH ALL THOSE NICE ARGUMENTS
	// -----

	if printVersion == true {
		fmt.Println("findnearest version", VERSION)
		os.Exit(0)
	}

	////
	if tgt == "" || univ == "" {
		printUsage()
		os.Exit(1)
	}

	fmt.Println("Target file: ", tgt)
	fmt.Println("Target file separator: ", tsep)
	fmt.Println("Universe file: ", univ)
	fmt.Println("Target file separator: ", usep)
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

	if tlat < 1 {
		fmt.Println("Please provide a positive integer as value for tlat.")
		os.Exit(1)
	}
	if tlng < 1 {
		fmt.Println("Please provide a positive integer as value for tlng.")
		os.Exit(1)
	}
	if ulat < 1 {
		fmt.Println("Please provide a positive integer as value for ulat.")
		os.Exit(1)
	}
	if ulng < 1 {
		fmt.Println("Please provide a positive integer as value for tlng.")
		os.Exit(1)
	}

	targetReader := csv.NewReader(targetFile)
	//targetReader.Comma = ';'                // Use semi-colon instead of comma
	targetReader.Comma = getSeparator(tsep) // Use semi-colon instead of comma
	//targetReader.FieldsPerRecord = -1 // number of expected fields ???
	targetData, err := targetReader.ReadAll()
	exitOnError(err, "Could not read target data")
	targetLatIndex := tlat - 1 // slice index starts at zero, tlat at 1
	targetLngIndex := tlng - 1 // slice index starts at zero, tlng at 1

	universeReader := csv.NewReader(universeFile)
	//universeReader.Comma = '\t' // Use tab instead of comma
	universeReader.Comma = getSeparator(usep)
	universeData, err := universeReader.ReadAll()
	exitOnError(err, "Could not read universe data")
	universeLatIndex := ulat - 1 // slice index starts at zero, ulat at 1
	universeLngIndex := ulng - 1 // slice index starts at zero, ulng at 1

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

// getSeparator converts a separator string to a rune
// Copied from https://github.com/spiffytech/csvmaster/blob/master/csvmaster.go
func getSeparator(sepString string) (sepRune rune) {
	sepString = `'` + sepString + `'`
	sepRunes, err := strconv.Unquote(sepString)
	if err != nil {
		if err.Error() == "invalid syntax" { // Single quote was used as separator. No idea why someone would want this, but it doesn't hurt to support it
			sepString = `"` + sepString + `"`
			sepRunes, err = strconv.Unquote(sepString)
			if err != nil {
				panic(err)
			}

		} else {
			panic(err)
		}
	}
	sepRune = ([]rune(sepRunes))[0]

	return sepRune
}

// printUsage prints usage output
//TODO see if this makes sense given automatic -h -help handling
func printUsage() {
	println("findnearest version 0.1-SNAPSHOT")
	println("")
	println("Usage:")
	flag.PrintDefaults()
	//println("")
	//println("Examples:")
	//println("  arrivahash -value 1 -salt 2")
	//println("  arrivahash -file \"/path/to/file\" -salt 2 > myOutput.txt")
}
