package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hailocab/go-geoindex"
	"github.com/paulmach/go.geo"
)

//TODO move initialization of flags to init function???
// see https://github.com/spiffytech/csvmaster/blob/master/csvmaster.go

//TODO rewrite to leverage spatial index - WIP

//VERSION is the version number of findnearest
const VERSION = "0.3.1"

//NAIVE  if set to false that will prevent execution of naive code
// naive code should be cleaned up if index works
const NAIVE = false

var (
	all = func(_ geoindex.Point) bool { return true }
)

func main() {

	// -----
	// ARGUMENTS
	// -----
	var tgt string
	var univ string
	var tlat, tlng int //tlat_index, tlng_index
	//also allow tlat_name, tlng_name ???
	var ulat, ulng int //ulat_index, ulng_index
	//also allow ulat_name, ulng_name ???
	var tsep, usep string
	var out string
	var printVersion bool
	//var printHelp bool

	//TODO let max distance be parameter
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

	flag.StringVar(&out, "out", "result.csv", "(Optional) Full path to output file")
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
	fmt.Println("Output file: ", out)

	targetFile, err := os.Open(tgt)
	exitOnError(err, "Oops, cannot find target file "+tgt)
	defer targetFile.Close()

	universeFile, err := os.Open(univ)
	exitOnError(err, "Oops, cannot find universe file "+univ)
	defer universeFile.Close()

	outputFile, err := os.Create(out)
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

	pointsIndex := geoindex.NewPointsIndex(geoindex.Km(0.5))

	if !NAIVE {
		for ukey, urecord := range universeData {
			if ukey < 1 {
				continue
			}
			l := len(urecord)
			if l <= maxUIndex {
				fmt.Println("Universe LatLng index out of range at line ", ukey+1, "")
				continue
			}

			//TODO check if we actually find lat lng coordinates at index
			universeLat, _ := strconv.ParseFloat(urecord[universeLatIndex], 64)
			universeLng, _ := strconv.ParseFloat(urecord[universeLngIndex], 64)

			//		up, err := newPointFromLatLngStrings(record[universeLatIndex], record[universeLngIndex])
			//		if err != nil {
			//			fmt.Println(err.Error())
			//			continue
			//		}

			upoint := &geoindex.GeoPoint{Pid: strconv.Itoa(ukey), Plat: universeLat, Plon: universeLng}
			pointsIndex.Add(upoint)
		}

		for tkey, trecord := range targetData {
			if tkey < 1 {
				continue
			}
			l := len(trecord)
			if l <= maxTIndex {
				fmt.Println("Target LatLng index out of range at line ", tkey+1, "")
				continue
			}

			//TODO check if we actually find lat lng coordinates at index
			targetLat, err := strconv.ParseFloat(trecord[targetLatIndex], 64)
			if err != nil {
				//TODO provide feedback that no lat coordinate was found
				continue
			}
			targetLng, err := strconv.ParseFloat(trecord[targetLngIndex], 64)
			if err != nil {
				//TODO provide feedback that no lng coordinates were found
				continue
			}

			tpoint := &geoindex.GeoPoint{Pid: strconv.Itoa(tkey), Plat: targetLat, Plon: targetLng}
			//TODO allow N nearest vs just nearest
			// but how do we save this in output ??? prefix columns from universe ???
			//TODO what happens if KNearest does not return any point
			//NOTE this is probably faster if Km is lower (add this as parameter)
			nearest := pointsIndex.KNearest(tpoint, 1, geoindex.Km(999.0), all)
			nPoint := nearest[0]
			//TODO check if nPoint is a point, otherwise continue
			nID := nPoint.Id()
			//since we add the slice index as id to GeoPoint for universe records
			// Atoi should never return an error if we have at least one result for nearest
			uIndex, _ := strconv.Atoi(nID)
			uRecord := universeData[uIndex]
			result := append(trecord, uRecord...)
			results = append(results, result)
		}

		csvWriter := csv.NewWriter(outputFile)
		csvWriter.WriteAll(results)
	}

	if NAIVE {
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
	println("findnearest version " + VERSION)
	println("")
	println("Usage:")
	// Some dependency imports testing package in non test files, because printDefaults prints all test flags
	// https://www.gmarik.info/blog/2016/go-testing-package-side-effects/
	flag.PrintDefaults()
	//println("")
	//println("Examples:")
	//println("  arrivahash -value 1 -salt 2")
	//println("  arrivahash -file \"/path/to/file\" -salt 2 > myOutput.txt")
}
