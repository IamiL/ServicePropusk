package buildinghandler

import (
	"fmt"
	"io"
	"net/http"
	buildService "rip/internal/service/build"
)

func AddBuildingPreview(buildingsService *buildService.BuildingService) func(
	w http.ResponseWriter,
	r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		// 32 MB is the default used by FormFile() function
		if err := r.ParseMultipartForm(10); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get a reference to the fileHeaders.
		// They are accessible only after ParseMultipartForm is called
		files := r.MultipartForm.File["file"]

		//var errNew string
		//var http_status int

		if len(files) != 1 {
			http.Error(w, "file not exists", http.StatusBadRequest)

			return
		}
		// Open the file
		file, err := files[0].Open()
		if err != nil {
			//errNew = err.Error()
			//http_status = http.StatusInternalServerError
			http.Error(w, "", http.StatusBadRequest)

			return
		}
		fmt.Println("1")
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			// Если не удается прочитать содержимое файла, возвращаем ошибку с соответствующим статусом и сообщением
			http.Error(w, "", http.StatusBadRequest)
			fmt.Println("err in fileBytes, err := io.ReadAll(file)")
			return
		}

		//buff := make([]byte, 512)
		//_, err = file.Read(buff)
		//if err != nil {
		//	//errNew = err.Error()
		//	//http_status = http.StatusInternalServerError
		//
		//	http.Error(w, "", http.StatusBadRequest)
		//
		//	return
		//}
		fmt.Println("2")

		// checking the content type
		// so we don't allow files other than images
		//filetype := http.DetectContentType(buff)
		//if filetype != "image/jpeg" && filetype != "image/png" && filetype != "image/jpg" {
		//	//errNew = "The provided file format is not allowed. Please upload a JPEG,JPG or PNG image"
		//	//http_status = http.StatusBadRequest
		//
		//	http.Error(w, "", http.StatusBadRequest)
		//
		//	return
		//}
		fmt.Println("3")

		//_, err = file.Seek(0, io.SeekStart)
		//if err != nil {
		//	//errNew = err.Error()
		//	//http_status = http.StatusInternalServerError
		//
		//	http.Error(w, "", http.StatusInternalServerError)
		//
		//	return
		//}
		//fmt.Println("4")

		if err := buildingsService.EditBuildingPreview(
			r.Context(),
			id,
			fileBytes,
		); err != nil {
			http.Error(w, "", http.StatusInternalServerError)

			return
		}

		//err = os.MkdirAll("./uploads", os.ModePerm)
		//if err != nil {
		//	errNew = err.Error()
		//	http_status = http.StatusInternalServerError
		//}
		//
		//f, err := os.Create(
		//	fmt.Sprintf(
		//		"./uploads/%d%s",
		//		time.Now().UnixNano(),
		//		filepath.Ext(files[0].Filename),
		//	),
		//)
		//if err != nil {
		//	errNew = err.Error()
		//	http_status = http.StatusBadRequest
		//}
		//
		//defer f.Close()
		//
		//_, err = io.Copy(f, file)
		//if err != nil {
		//	errNew = err.Error()
		//	http_status = http.StatusBadRequest
		//}
		//message := "file uploaded successfully"
		//messageType := "S"
		//
		//if errNew != "" {
		//	message = errNew
		//	messageType = "E"
		//}
		//
		//if http_status == 0 {
		//	http_status = http.StatusOK
		//}

		//resp := map[string]interface{}{
		//	"messageType": messageType,
		//	"message":     message,
		//}
		//w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		//json.NewEncoder(w).Encode(resp)
	}
}
