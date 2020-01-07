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
	geo "github.com/paulmach/go.geo"
)

//TODO move initialization of flags to init function???
// see https://github.com/spiffytech/csvmaster/blob/master/csvmaster.go

//VERSION is the version number of findnearest
const VERSION = "0.4"

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
	var calcDist bool
	var printVersion bool
	var printHelp bool

	flag.StringVar(&tgt, "target", "", "(Required) Path to target file.")
	flag.StringVar(&univ, "universe", "", "(Required) Path to universe file.")
	flag.IntVar(&tlat, "tlat", 0, "(Required) Index of Latitude column in target file.")
	flag.IntVar(&tlng, "tlng", 0, "(Required) Index of Latitude column in target file.")

	flag.IntVar(&ulat, "ulat", 0, "(Required) Index of Latitude column in universe file.")
	flag.IntVar(&ulng, "ulng", 0, "(Required) Index of Longitude column in universe file.")

	flag.StringVar(&tsep, "tsep", ",", "(Optional) Field separator in target file ('tab' for tab-separated).")
	flag.StringVar(&usep, "usep", ",", "(Optional) Field separator in universe file ('tab' for tab-separated).")

	flag.StringVar(&out, "out", "result.csv", "(Optional) Full path to output file")
	flag.BoolVar(&calcDist, "dist", true, "Set -dist=false to disable the addition of a distance column. Defaults to true.")
	flag.BoolVar(&printVersion, "version", false, "Print program version")
	flag.BoolVar(&printHelp, "h", false, "Print help")

	flag.Usage = func() {
		printUsage()
	}

	flag.Parse()

	// -----
	// DO STH WITH ALL THOSE NICE ARGUMENTS
	// -----

	if printVersion == true {
		fmt.Println("findnearest version", VERSION)
		os.Exit(0)
	}

	if printHelp == true {
		printUsage()
		os.Exit(0)
	}

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
	targetReader.Comma = getSeparator(tsep)
	targetData, err := targetReader.ReadAll()
	exitOnError(err, "Could not read target data")
	targetLatIndex := tlat - 1 // slice index starts at zero, tlat at 1
	targetLngIndex := tlng - 1 // slice index starts at zero, tlng at 1

	universeReader := csv.NewReader(universeFile)
	universeReader.Comma = getSeparator(usep)
	universeData, err := universeReader.ReadAll()
	exitOnError(err, "Could not read universe data")
	universeLatIndex := ulat - 1 // slice index starts at zero, ulat at 1
	universeLngIndex := ulng - 1 // slice index starts at zero, ulng at 1

	var results [][]string

	maxTIndex := max(targetLatIndex, targetLngIndex)
	maxUIndex := max(universeLatIndex, universeLngIndex)
	header := append(targetData[0], universeData[0]...)
	// add an extra column for the distance between target and universe points
	if calcDist {
		header = append(header, "Distance")
	}
	results = append(results, header)

	pointsIndex := geoindex.NewPointsIndex(geoindex.Km(0.5))

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

		targetLat, err := strconv.ParseFloat(trecord[targetLatIndex], 64)
		if err != nil {
			//Provide feedback that no lat coordinate was found?
			continue
		}
		targetLng, err := strconv.ParseFloat(trecord[targetLngIndex], 64)
		if err != nil {
			//Provide feedback that no lng coordinates were found?
			continue
		}

		tpoint := &geoindex.GeoPoint{Pid: strconv.Itoa(tkey), Plat: targetLat, Plon: targetLng}
		//NOTE this may be faster if Km is lower. We could add Km as a parameter)
		nearest := pointsIndex.KNearest(tpoint, 1, geoindex.Km(999.0), all)
		nPoint := nearest[0]
		//TODO check if nPoint is a point, otherwise continue
		nID := nPoint.Id()
		//since we add the slice index as id to GeoPoint for universe records
		// Atoi should never return an error if we have at least one result for nearest
		uIndex, _ := strconv.Atoi(nID)
		uRecord := universeData[uIndex]
		tuDistance := geoindex.Distance(tpoint, nPoint)
		result := append(trecord, uRecord...)
		// calculate and add the distance to the result file in the last column
		if calcDist {
			result = append(result, fmt.Sprintf("%f", tuDistance))
		}
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
func printUsage() {
	println("findnearest version " + VERSION)
	println("")
	println("Usage:")
	// Some dependency imports testing package in non test files, because printDefaults prints all test flags
	// https://www.gmarik.info/blog/2016/go-testing-package-side-effects/
	//flag.PrintDefaults()
	println("-dist")
	println("    Set -dist=false to disable the addition of a distance column. Defaults to true. ")
	println("-h")
	println("    Show help")
	println("-out")
	println("    (Optional) Path to output file")
	println("-target <path>")
	println("    (Required) Path to target file.")
	println("-tlat <index>")
	println("    (Required) Index of Latitude column in target file.")
	println("-tlng <index>")
	println("    (Required) Index of Longitude column in target file.")
	println("-tsep <sep>")
	println("    (Optional) Field separator in target file ('tab' for tab-separated).")
	println("-universe <path>")
	println("    (Required) Path to universe file.")
	println("-ulat <index>")
	println("    (Required) Index of Latitude column in universe file.")
	println("-ulng <index>")
	println("    (Required) Index of Longitude column in universe file.")
	println("-usep <sep>")
	println("    (Optional) Field separator in universe file ('tab' for tab-separated).")
	println("")
	println("Example:")
	println("findnearest -target data/target.txt -tlat 2 -tlon 3 -tsep ; -universe data/universe.txt -ulat 2 -ulon 3 -usep ;")
}
