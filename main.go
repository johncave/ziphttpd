package main

import (
    "os"
    "fmt"
    "net/http"
    "strings"
    "mime"
    "strconv"
    "io/ioutil"
    "compress/gzip"
    "bytes"
    "code.google.com/p/gcfg"
    "time"
)

// Everything we expect to find in the config file.
type conf struct{
  Global struct{
    Webroot string
    Ziproot string
    Expires int
    CachePragmas string
    Port string
    AllowDirlist bool
    SSLPort string
    SSLCert string
    SSLKey string
    CreateFiles bool
    TypesToZip []string
  }
 
}
// All global configuration is stored in "config" of type "conf". 
// Confusing, but it made sense at the time O.o
var config conf // Global configuration.


// Handler handles all incoming connections first.
func handler(w http.ResponseWriter, r *http.Request) {
// General pre-serve stuff.
    //fmt.Print("Recieved a request. ")
    var file string = config.Global.Webroot+r.URL.Path[1:] 
    w.Header().Set("Server", "ziphttpd")
    w.Header().Set("Cache-Control", "max-age="+strconv.Itoa(config.Global.Expires)+", "+config.Global.CachePragmas)

// Work on serving the file.
    if(exists(file, config.Global.AllowDirlist)){
	if(strings.Contains( r.Header.Get("Accept-Encoding"), "gzip")){ //If client reports accepting Gzip
	    handlerGzip( w, r, file)
	} else { // If client can't handle the truth.
	    http.ServeFile(w, r, file)
	}
	//fmt.Println("It existed so I served it.")
    } else { // File doesn't exist.
	http.Error(w, "Sorry, this file could not be found: "+r.URL.Path, http.StatusNotFound)
	//fmt.Println("It didn't exist so I served an error")
    }    


}

// Handles the request if the client supports gzip encoding.
// Maybe: Add support to gzip files on the fly if pre-gzipped version is not found.
func handlerGzip(w http.ResponseWriter, r *http.Request, file string){
//fmt.Print("Entered Gzip method. ")
    if(exists(config.Global.Ziproot+r.URL.Path[1:]+".gz", false)){ // If gzipped file exists.
	// Figure out the file type and set that header.
	bigfile, err := os.Stat(file)
	zipfile, err := os.Stat(config.Global.Ziproot+r.URL.Path[1:]+".gz")
	//fmt.Println("Zipfile modtime is: ", zipfile.ModTime().Unix(), " Bigfile modtime is: ", bigfile.ModTime().Unix())
	if(err == nil && zipfile.ModTime().Unix() != bigfile.ModTime().Unix()){
	  createGzip(file, r.URL.Path[1:])
	  //fmt.Println("Gzip file does not have same modified time")
	}
	var fileType []string = strings.Split(file, ".")
	w.Header().Set( "Content-Type", mime.TypeByExtension("."+fileType[len(fileType)-1]))
	w.Header().Set( "Content-Encoding", "gzip") // Tell the browser the content is zipped.
	// Serve the file.
	http.ServeFile(w, r, config.Global.Ziproot+r.URL.Path[1:]+".gz")
	//fmt.Println("Served .gz file at "+config.Global.Ziproot+r.URL.Path[1:]+".gz")
    } else { // Decide whether to serve the file uncompressed, or compress then serve it.
	var fileType []string = strings.Split(file, ".")
	//fmt.Print("Is "+mime.TypeByExtension("."+fileType[len(fileType)-1])+" cacheable? ")
	//fmt.Println(stringContains(mime.TypeByExtension("."+fileType[len(fileType)-1]), config.Global.TypesToZip))
	
	if(stringContains(mime.TypeByExtension("."+fileType[len(fileType)-1]), config.Global.TypesToZip) && config.Global.CreateFiles){ // If the MIME type contains "text" we assume that compressing it will be worth our time.
	  if(createGzip(file, r.URL.Path[1:])) { // If the gzip file was created.
	    
	    handlerGzip(w, r, file)
	    //fmt.Println("Created gz file and served it.")
	  } else{
	    http.ServeFile(w, r, file)
	    //fmt.Println("Couldn't create gz file, so served normal version.")
	  }
	
	} 	else{
	  http.ServeFile(w, r, file)
	  //fmt.Println("Served normal file because this file cannot be compressed.")
	}
    }
}


// Exists returns whether the given file exists or not. A directory is not counted as a file in this instance.
// Maybe: Add support to turn directory listing on and off.
// To do: Nothing at the moment.
func exists(path string, dirs bool) (bool) {
    file, err := os.Stat(path)
    if(dirs && err == nil){ 
      return true
      } else if( err == nil) { 
	return !file.IsDir()
    }
    if os.IsNotExist(err) { return false }
    return false
}

//Same as above, but to determine if directory exists or not.
func dirExists(path string) (bool) {
    _, err := os.Stat(path)
    if err == nil { 
	return true
    }
    if os.IsNotExist(err) { return false }
    return false
}

// Creates a Gzip of a file, if it can, and stores it.
func createGzip(file string, path string)(bool){

    // See if target directory exists before creating .gz file.
    var dir []string
    dir = strings.Split(config.Global.Ziproot+path, "/")
    dir = dir[0:len(dir)-1]
    var destDir string
    destDir = strings.Join(dir, "/")
    if(!dirExists(destDir)){
     os.MkdirAll(destDir, 0700) 
      
    }
    
    // Create file
    fileContents, error := ioutil.ReadFile(file)
    var buffer bytes.Buffer
    writer, _ := gzip.NewWriterLevel(&buffer, gzip.BestCompression) // For some reason, this writer has to be stored in a var.
    writer.Write(fileContents)
    writer.Close()
    error = ioutil.WriteFile(config.Global.Ziproot+path+".gz", buffer.Bytes(), 0700)
    if(error != nil){
      return false // Something has gone wrong and the .gz file could not be created. Serve normal file.
      fmt.Println(error)
    }
    fileStats, err := os.Stat(file)
    err = os.Chtimes(config.Global.Ziproot+path+".gz", time.Now().Local(), fileStats.ModTime())
    if(err != nil){
      fmt.Println(error)
    }
    //fmt.Println("Probably made gz file.")
    return true
    
}


func stringContains(value string, list []string) bool {
	for _, v := range list {
		if strings.Contains(value, v) {
			return true
		}
	}
	return false
}

func main() {
    //Import Config file.    
    var err error
    if(len(os.Args)== 0){
      err := gcfg.ReadFileInto(&config, os.Args[1])
      if(err != nil){
	panic(err)
      }
    } else {
      err := gcfg.ReadFileInto(&config, "ziphttpd.cfg")
      if(err != nil){
	panic(err)
      }
    } 
    fmt.Println("Read config file.")
    if(err != nil){
      panic(err)
    }
    
    // Welcome the user to Ziphttpd.    
    fmt.Println("       ^_^\n       \\|/\n	|\n       / \\\nZiphttpd is started. Yay!")
    
    //Start Serving.
    http.HandleFunc("/", handler)
    if(len(config.Global.SSLPort) > 0){
      go http.ListenAndServeTLS(config.Global.SSLPort, config.Global.SSLCert, config.Global.SSLKey, nil)
      fmt.Println("Listening HTTPS")
    }
    http.ListenAndServe(config.Global.Port, nil)
    fmt.Println("Listening HTTP")
}
