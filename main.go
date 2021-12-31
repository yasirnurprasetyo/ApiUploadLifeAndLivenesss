package main

import (
	"face-liveness-privy/pkg/ffmpeg"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lijo-jose/gffmpeg/pkg/gffmpeg"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Configuration defines the settings required during the app initiation
type Configuration struct {
	inputFilePath string
	destination   string
}
 
const (
	inputFilePathArg = "input"
	destinationArg   = "dest"
)
 
func init() {
 
	pflag.String(inputFilePathArg, "upload", "input file for processing")
	pflag.String(destinationArg, "datasets/", "destination directory for storing results and intermediate files")
 
	viper.AutomaticEnv()
	viper.BindPFlag(inputFilePathArg, pflag.Lookup(inputFilePathArg))
	viper.BindPFlag(destinationArg, pflag.Lookup(destinationArg))
 
	pflag.Parse()
 
}
 
func UploadFile(svc ffmpeg.Service, cfg Configuration) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
 
		//limit  10MB
		request.ParseMultipartForm(10 * 1024 * 1024)
 
		file, handler, err := request.FormFile("myFile")
		if err != nil {
			fmt.Println(err)
			return
		}
 
		defer file.Close()
 
		fmt.Println("File Info")
		fmt.Println("File Name", handler.Filename)
		fmt.Println("File Size", handler.Size)
		fmt.Println("File Type", handler.Header.Get("Content-Type"))
 
		test := handler.Filename
 
		fmt.Println(test)
 
		//Upload file
 
		tempFile, err2 := ioutil.TempFile("upload", "upload-*.mp4")
		if err != nil {
			fmt.Println(err2)
		}
		defer tempFile.Close()
 
		fileBytes, err3 := ioutil.ReadAll(file)
		if err3 != nil {
			fmt.Println(err3)
		}
		tempFile.Write(fileBytes)
		fmt.Fprintf(response, "Upload successful")
 
		namefile := tempFile.Name()
		nf := filepath.Base(namefile)
		dest := strings.TrimSuffix(nf, filepath.Ext(nf))
		//Create a folder/directory at a full qualified path
		os.MkdirAll(cfg.destination+dest, 0755)
 
		err = svc.ExtractFrames(namefile, cfg.destination+dest+"/", 1)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := ioutil.ReadFile("/datasets/upload-489729458/frames1.jpg")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
}
 
func main() {
	ff, err := gffmpeg.NewGFFmpeg("/usr/bin/ffmpeg")
	if err != nil {
		fmt.Println(err)
 
		return
	}
 
	svc, err := ffmpeg.New(ff)
	if err != nil {
		fmt.Println(err)
 
		return
	}
 
	router := mux.NewRouter()
 
	router.HandleFunc("/api/upload", UploadFile(svc, Configuration{
		inputFilePath: viper.GetString(inputFilePathArg),
		destination:   viper.GetString(destinationArg),
	})).Methods("POST")
	router.HandleFunc("/api/gambar", handleRequest).Methods("GET")
 
	err = http.ListenAndServe(":8001", router)
	if err != nil {
		fmt.Println(err)
	}
}