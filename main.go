package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("example-web-server")

	err := ioutil.WriteFile("./file-1.txt", []byte("File 1"), os.ModePerm)
	check(err)
	err = ioutil.WriteFile("./file-2.txt", []byte("File 2"), os.ModePerm)
	check(err)
	files := []string{"file-1.txt", "file-2.txt"}
	output := "done.zip"

	err = ZipFiles(output, files)
	check(err)

	dat, err := ioutil.ReadFile(output)
	check(err)

	fmt.Println("Zipped File:", output)

	StartServer("localhost:8080", []byte(dat))

}

type downloadHandler struct {
	doneSignalChan chan<- bool
	report         []byte
}

func (h downloadHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("download") != "" {
		rw.Header().Add("Content-Type", "application/zip")
		_, err := rw.Write(h.report)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Server error: %s", err.Error())
		}

		// h.doneSignalChan <- true
		return
	}
	// rw.Header().Add("Content-Type", "application/zip")

	_, err := rw.Write([]byte(`<html><body><a href="?download=report">Click here to download report.</a></body></html>`))
	if err != nil {
		fmt.Printf("Server error: %s", err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
	}
}

func StartServer(serverAddr string, report []byte) {
	// go func() {
	handler := downloadHandler{
		doneSignalChan: nil,
		report:         report,
	}

	server := &http.Server{Addr: serverAddr, Handler: handler}
	log.Fatal(server.ListenAndServe())
	// fmt.Printf("Error starting embedded webserver: %s", err)
	// }()
}

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = AddFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
