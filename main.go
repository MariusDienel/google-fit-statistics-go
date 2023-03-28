package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ExportFileName = "GoogleFitExport.zip"
	TempFolderName = "google-fit-tmp"
)

var (
	StartingDate = time.Date(2019, time.April, 1, 0, 0, 0, 0, time.UTC)
	EndDate      = time.Date(2023, time.October, 31, 0, 0, 0, 0, time.UTC)
)

func main() {
	tempPath := unzipFiles()
	bikeFiles := getFilesOfType(tempPath, "Radfahren")
	//fmt.Println(bikeFiles)

	distancesInM := make([]float64, 0)
	for _, file := range bikeFiles {
		distanceInM, err := getDistanceInM(file)
		if err != nil {
			log.Fatal(err)
		}
		distancesInM = append(distancesInM, distanceInM)
	}

	result := fmt.Sprintf("%.2f", sum(distancesInM))
	message := fmt.Sprintf("Du bist insgesamt %sm mit dem Rad gefahren", result)
	fmt.Println(message)

	cleanUpTempDir()
}

func cleanUpTempDir() {
	tempDirPath := getPathToSearch() + string(os.PathSeparator) + TempFolderName
	err := os.RemoveAll(tempDirPath)
	if err != nil {
		panic(err)
	}
}

func unzipFiles() string {
	zipPath := getPathToSearch() + string(os.PathSeparator) + ExportFileName
	zipReader, err := zip.OpenReader(zipPath)
	panicIfNotNull(err)
	defer zipReader.Close()

	tempPath := getPathToSearch() + string(os.PathSeparator) + TempFolderName
	os.MkdirAll(tempPath, os.ModePerm)
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, ".tcx") {
			copyFileToTemp(tempPath, file)
		}
	}
	return tempPath
}

func copyFileToTemp(tempPath string, file *zip.File) {
	split := strings.Split(file.Name, "/")
	filename := split[len(split)-1]
	dstFile, err := os.Create(filepath.Join(tempPath, filename))
	panicIfNotNull(err)
	sourceFromArchive, err := file.Open()
	panicIfNotNull(err)
	io.Copy(dstFile, sourceFromArchive)
	dstFile.Close()
	sourceFromArchive.Close()
}

func panicIfNotNull(err error) {
	if err != nil {
		panic(err)
	}
}

func getFilesOfType(tempPath string, fileType string) []string {
	files, err := os.ReadDir(tempPath)
	logError(err)

	var validFiles []string
	for _, file := range files {
		if strings.Contains(file.Name(), fileType) && hasValidDate(file.Name()) {
			validFiles = append(validFiles, file.Name())
		}
	}

	return validFiles
}

func logError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func hasValidDate(filename string) bool {
	fileDateString := filename[0:10]
	fileDate, err := time.Parse("2006-01-02", fileDateString)
	if err != nil {
		log.Fatal(err)
	}

	return StartingDate.Unix() < fileDate.Unix() && fileDate.Unix() < EndDate.Unix()
}

func getDistanceInM(filename string) (float64, error) {
	file, err := os.Open(getPathToSearch() + string(os.PathSeparator) + TempFolderName + string(os.PathSeparator) + filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
	}
	defer file.Close()

	// Read the file contents
	contents, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	// Parse the XML
	var data TrainingCenterDatabase
	err = xml.Unmarshal(contents, &data)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
	}

	var distances []float64
	for _, activity := range data.Activities.Activity {
		for _, lap := range activity.Lap {
			for _, trackpoint := range lap.Track {
				distances = append(distances, trackpoint.DistanceMeters)
			}
		}
	}
	max := maxFloat(distances)
	fmt.Printf("Max value for '%s' is '%f'", filename, max)
	fmt.Println("-------")
	return max, err
}

func maxFloat(numbers []float64) float64 {
	max := numbers[0]
	for i := 1; i < len(numbers); i++ {
		if numbers[i] > max {
			max = numbers[i]
		}
	}
	return max
}

func sum(numbers []float64) float64 {
	var sum float64
	for _, num := range numbers {
		sum += num
	}
	return sum
}

func getPathToSearch() string {
	userHome, err := os.UserHomeDir()
	logError(err)
	return userHome + string(os.PathSeparator) + "Downloads" + string(os.PathSeparator)
}

type Trackpoint struct {
	DistanceMeters float64 `xml:"DistanceMeters"`
}

type Lap struct {
	Track []Trackpoint `xml:"Track>Trackpoint"`
}

type Activity struct {
	Lap []Lap `xml:"Lap"`
}

type Activities struct {
	Activity []Activity `xml:"Activity"`
}

type TrainingCenterDatabase struct {
	Activities Activities `xml:"Activities"`
}
