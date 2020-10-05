package main

import "fmt"
import "net/http"
import "html/template"
import "os"
import "io"
import "path/filepath"
import "encoding/json"

type M map[string]interface{}

func handleIndex(w http.ResponseWriter, r *http.Request){
	tmpl := template.Must(template.ParseFiles("index.html"))
	if err := tmpl.Execute(w, nil);
	err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
		http.Error(w, "Only accept POST method", http.StatusBadRequest)
		return
	}

	basePath, _ := os.Getwd()
	reader, err := r.MultipartReader() //request milik handler
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for {
		part, err := reader.NextPart() //untuk mengembalikan 2 informasi
		if err == io.EOF {	
			break
		}
		fileLocation := filepath.Join(basePath, "file", part.FileName())
		dst, err := os.Create(fileLocation)
		if dst != nil {
			defer dst.Close()
		}
		if err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := io.Copy(dst, part); //untuk mengisi data stream 
		err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte(`all files uploaded`))
}

func ListFiles(w http.ResponseWriter, r *http.Request){
	tmpl := template.Must(template.ParseFiles("view.html"))
	if err := tmpl.Execute(w, nil);
	err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleListFiles(w http.ResponseWriter, r *http.Request){
	files := []M{}
	basePath,_ := os.Getwd() //mengembalikan informasi absolute path dimana aplikasi di-eksekusi.
	filesLocation := filepath.Join(basePath, "file") //digabung folder

	err := filepath.Walk(filesLocation, func(path string, info os.FileInfo, err error) error{
		//untuk membaca isi dari sebuah direktori, apa yang ada didalamnya (file maupun folder) akan di-loop. 
		if err != nil {
			return err 
		}
		if info.IsDir(){
			return nil
		}
		files = append(files, M{"filename": info.Name(), "path": path})
		return nil 
	})
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(files)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type","application/json")
	w.Write(res)
}

func handleDownload(w http.ResponseWriter, r *http.Request){
	if err := r.ParseForm();
	err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	path := r.FormValue("path")
	f, err := os.Open(path)
	if f != nil{
		defer f.Close()
	}
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//menghasilkan keluaran download pada browser 
	contentDisposition := fmt.Sprintf("attachment; filename=%s", f.Name()) 
	//salah satu protocol untuk menginformasikan browser untuk berinteraksi dengan output
	w.Header().Set("Content-Disposition", contentDisposition) //untuk menghubungkan dengan download 

	if _, err := io.Copy(w, f); //mengcopy variable f
	err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func main(){
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/main.go/listfile", ListFiles)
	http.HandleFunc("/list-files", handleListFiles)
	http.HandleFunc("/download", handleDownload)

	fmt.Println("server starting at localhost :8000")
	http.ListenAndServe(":8000", nil)
}