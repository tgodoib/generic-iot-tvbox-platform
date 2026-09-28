package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"ota_server/model"
	"ota_server/ota"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type HTTP struct {
	ota    ota.Service
	router *chi.Mux
}

func NewHTTPAdapter(ota ota.Service) *HTTP {

	/*
		POST   /firmware          # upload do binário, retorna 201 + { id, created_at }
		GET    /firmware/:id      # metadados do firmware (versão, tamanho, created_at)
		DELETE /firmware/:id      # remove firmware do Minio

		POST   /update/:id        # envia manifest via MQTT, retorna 202 Accepted
	*/

	r := chi.NewRouter()

	return &HTTP{
		ota:    ota,
		router: r,
	}
}

func (h *HTTP) Listen(addr string) {
	h.router.Use(middleware.Logger)

	h.router.Get("/firmware", h.ListFirmware)
	h.router.Post("/firmware", h.UploadFirmware)

	h.router.Get("/firmware/*", h.GetFirmware)
	h.router.Delete("/firmware/*", h.DeleteFirmware)

	h.router.Post("/firmware/update/*", h.SendUpdate)

	go func() {
		log.Fatal(http.ListenAndServe(addr, h.router))
	}()

	fmt.Println("HTTP server started on ", addr)
}

func (h *HTTP) ListFirmware(w http.ResponseWriter, r *http.Request) {
	list, err := h.ota.ListFirmware()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	payload, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

func (h *HTTP) UploadFirmware(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing file"))
		return
	}
	defer file.Close()

	version := r.FormValue("version")
	deviceClass := r.FormValue("device_class")

	if version == "" || deviceClass == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing version or device_class"))
		return
	}

	id, err := h.ota.CreateFirmware(model.Firmware{
		FirmwareFile: model.FirmwareFile{
			Bin:  file,
			Size: header.Size,
		},
		FirmwareMeta: model.FirmwareMeta{
			Version:     version,
			DeviceClass: deviceClass,
		},
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}

func (h *HTTP) GetFirmware(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "*")

	if id == "" {
		h.ListFirmware(w, r)
		return
	}

	data, err := h.ota.GetFirmwareData(id)

	fmt.Println(id, data, err)

	if err != nil {
		if errors.Is(err, model.ErrFirmwareNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

func (h *HTTP) DeleteFirmware(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "*")

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := h.ota.DeleteFirmware(id)

	if err != nil {
		if errors.Is(err, model.ErrFirmwareNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTP) SendUpdate(w http.ResponseWriter, r *http.Request) {
	var recipient model.Recipient

	if err := json.NewDecoder(r.Body).Decode(&recipient); err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	id := chi.URLParam(r, "*")

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := h.ota.SendUpdate(id, recipient)

	if err != nil {
		if errors.Is(err, model.ErrFirmwareNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
