package handler

//AQUI ESTARAN LOS RUTEOS Y LOS MIDDLWARE

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/IgnacioBO/go_lib_response/response"
	"github.com/IgnacioBO/gomicro_enrollment/internal/enrollment"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// Este recibe contexto y un endpoint que definimos en la capa del endpionts
func NewUserHTTPServer(ctx context.Context, endpoints enrollment.Endpoints) http.Handler {
	router := mux.NewRouter()

	//Esta se guarad en opciones y se pone al final en Handle
	opciones := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}

	//Ahora usaremos Handle, poreque a este se le puede pasar un server (httptranpsort)
	router.Handle("/enrollments", httptransport.NewServer(
		endpoint.Endpoint(endpoints.Create),
		decodeCreateEnrollment,
		encodeResponse,
		opciones...,
	)).Methods("POST")

	router.Handle("/enrollments", httptransport.NewServer(
		endpoint.Endpoint(endpoints.GetAll),
		decodeGetAllEnrollment,
		encodeResponse,
		opciones...,
	)).Methods("GET")

	router.Handle("/enrollments/{id}", httptransport.NewServer(
		endpoint.Endpoint(endpoints.Update),
		decodeUpdateEnrollment,
		encodeResponse,
		opciones...,
	)).Methods("PATCH")

	return router
}

// *** MIDDLEWARE REQUEST ***
func decodeCreateEnrollment(_ context.Context, r *http.Request) (interface{}, error) {
	var reqStruct enrollment.CreateRequest

	//Ahora hacemos el decode del body del json al srtuct de REquest de course
	err := json.NewDecoder(r.Body).Decode(&reqStruct)
	if err != nil {
		return nil, response.BadRequest(fmt.Sprintf("invalid request format: '%v'", err.Error()))
	}

	return reqStruct, nil
}

// *** MIDDLEWARE RESPONSE ***
func encodeResponse(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	rInterface := resp.(response.Response)                            //Transformamos el resp a response.Respone (al interface) -> YA QUE LE ENAIREMOS SIEMPRE UN objeto RESPONSE (CREADO POR NOSOTROS, q tiene el code, mensage, meta, etc, todo el json)
	w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json
	w.WriteHeader(rInterface.StatusCode())
	return json.NewEncoder(w).Encode(rInterface) //resp tendra el user.User del domain y otroas datos si es necesario para ocnveritse en json

}

// *** MIDDLEWARE RESPONSE DE ERROR ***
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json
	respInterface := err.(response.Response)                          //Tranfosrmamos el error recibido a la interfac response.Response que craemos
	//¿Porque funciona esta conversion de tipo error al de nosotros?, porque la interfaz 'error' de go pide que haya un metodo Error() string [QUE CREAMOS EN nuestro respon.RESPONSE!]
	//Entonces como implementamos el metodo Error() string funcinoa, ademas tenemos al ventaja que vamos apoder obtener MAS DATOS porque repsonse.Response tiene mas metodos como (StatusCode())
	//Entonces podemos transofrmar un error a una interfac propia con MAS METODOS Y MAS DATOS UE UN ERROR NORMAL!
	w.WriteHeader(respInterface.StatusCode())
	_ = json.NewEncoder(w).Encode(respInterface)

}
func decodeGetAllEnrollment(_ context.Context, r *http.Request) (interface{}, error) {
	fmt.Println("decode getall enroll")
	variablesURL := r.URL.Query()

	//Ahora obtendremos el limit y la pagina desde los parametros
	limit, _ := strconv.Atoi(variablesURL.Get("limit"))
	page, _ := strconv.Atoi(variablesURL.Get("page"))

	getReqAll := enrollment.GetAllRequest{
		UserID:   variablesURL.Get("user_id"),
		CourseID: variablesURL.Get("course_id"),
		Limit:    limit,
		Page:     page,
	}

	return getReqAll, nil
}

func decodeUpdateEnrollment(_ context.Context, r *http.Request) (interface{}, error) {
	var reqStruct enrollment.UpdateRequest

	err := json.NewDecoder(r.Body).Decode(&reqStruct)
	if err != nil {
		return nil, response.BadRequest(fmt.Sprintf("invalid request format: '%v'", err.Error()))
	}

	variablesPath := mux.Vars(r)
	reqStruct.ID = variablesPath["id"]

	return reqStruct, nil

}
