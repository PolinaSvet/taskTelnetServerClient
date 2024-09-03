package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Программный интерфейс сервера GoNews
type API struct {
	router      *mux.Router
	sendChannel chan string
	dataChannel chan string
	upgrader    websocket.Upgrader
}

// Конструктор объекта API
func New(sendChannel, dataChannel chan string) *API {
	api := API{
		sendChannel: sendChannel,
		dataChannel: dataChannel,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	api.router = mux.NewRouter()
	api.endpoints()
	return &api
}

// Регистрация обработчиков API.
func (api *API) endpoints() {

	api.router.HandleFunc("/", api.templateHandler).Methods(http.MethodGet, http.MethodOptions)
	api.router.HandleFunc("/ws", api.wsHandler)

	api.router.HandleFunc("/sendDataTo", api.sendDataToHandler).Methods(http.MethodPost, http.MethodOptions)

	// Регистрация обработчика для статических файлов (шаблонов)
	api.router.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))))
}

// Получение маршрутизатора запросов.
// Требуется для передачи маршрутизатора веб-серверу.
func (api *API) Router() *mux.Router {
	return api.router
}

// Обработчик веб-сокетов
func (api *API) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error while upgrading connection:", err)
		return
	}
	defer conn.Close()

	for data := range api.dataChannel {
		err := conn.WriteMessage(websocket.TextMessage, []byte(data))
		if err != nil {
			log.Println("Error while writing message:", err)
			return
		}
	}
}

// Базовый маршрут.
func (api *API) templateHandler(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.ParseFiles("ui/html/base.html", "ui/html/routes.html"))

	// Отправляем HTML страницу с данными
	if err := tmpl.ExecuteTemplate(w, "base", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Расчет одного выражения.
func (api *API) sendDataToHandler(w http.ResponseWriter, r *http.Request) {
	//принимаем данные
	var jsonDataMap map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&jsonDataMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//преобразуем в нужный формат
	var line []string
	for _, v := range jsonDataMap {
		line = append(line, fmt.Sprintf("%v", v))
	}
	msg := "error msg..."
	if len(line) > 0 {
		msg = line[0]
	}
	api.sendChannel <- msg

	//отдаем данные
	bytes, err := json.Marshal(msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
