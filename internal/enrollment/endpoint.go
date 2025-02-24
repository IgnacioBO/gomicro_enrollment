package enrollment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/IgnacioBO/go_lib_response/response"
	"github.com/IgnacioBO/gomicro_meta/meta"

	courseSdk "github.com/IgnacioBO/go_micro_sdk/course"
	userSdk "github.com/IgnacioBO/go_micro_sdk/user"
)

type (
	//Controller sera una funcion que reciba REspone y Request
	Controller func(ctx context.Context, request interface{}) (interface{}, error)
	Endpoints  struct {
		Create Controller
		GetAll Controller
		Update Controller
	}
	//Definiremos una struct para definir el request del Craete, con los campos que quiero recibir y los tags de json
	CreateRequest struct {
		UserID   string `json:"user_id"`
		CourseID string `json:"course_id"`
	}

	UpdateRequest struct {
		ID     string  `json:"id"`
		Status *string `json:"status"`
	}
	//Struct para guardar la cant page por defecto y otras conf
	Config struct {
		LimitPageDefault string
	}

	GetAllRequest struct {
		UserID   string
		CourseID string
		Limit    int
		Page     int
	}
)

// Funcion que se encargará de hacer los endopints
// Para eso necesitaremos una struct que se llamara endpoints
// Esta funcion va a DEVOLVER una struct de Endpoints, estos endpoints son los que vamos a poder utuaizlar en unestro dominio (course)
func MakeEndpoints(s Service, c Config) Endpoints {
	return Endpoints{
		Create: makeCreateEndpoint(s),
		GetAll: makeGetAllEndpoint(s, c),
		Update: makeUpdateEndpoint(s),
	}
}

func makeCreateEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("create course")

		//Variable con struct de request (datos usaurio)
		reqStruct := request.(CreateRequest)
		//Validaciones
		if reqStruct.UserID == "" {
			return nil, response.BadRequest(ErrUserIDRequired.Error())
		}
		if reqStruct.CourseID == "" {
			return nil, response.BadRequest(ErrCourseIDRequired.Error())
		}

		fmt.Println(reqStruct)
		reqStrucEnJson, _ := json.MarshalIndent(reqStruct, "", " ")
		fmt.Println(string(reqStrucEnJson))

		enrollNuevo, err := s.Create(ctx, reqStruct.UserID, reqStruct.CourseID)
		if err != nil {
			if errors.As(err, &userSdk.ErrNotFound{}) || errors.As(err, &courseSdk.ErrNotFound{}) {
				return nil, response.NotFound(err.Error())
			}
			return nil, response.InternalServerError(err.Error())
		}

		return response.Created("success", enrollNuevo, nil), nil
	}
}

func makeGetAllEndpoint(s Service, config Config) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("getall user")

		getAllParametros := request.(GetAllRequest)
		//Luego con podemos acceder a los parametos y guardarlos en el struct Filtro (creado en service.go)
		filtros := Filtros{
			UserID:   getAllParametros.UserID,
			CourseID: getAllParametros.CourseID,
		}

		//Ahora llamaremos al Count del service que creamos (antes de hacer la consulta completa)
		cantidad, err := s.Count(ctx, filtros)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}
		//Luego crearemos un meta y le agregaremos la cantidad que consultamos, luego el meta lo ageregaremos a la respuesta
		meta, err := meta.New(getAllParametros.Page, getAllParametros.Limit, cantidad, config.LimitPageDefault)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}

		allUsers, err := s.GetAll(ctx, filtros, meta.Offset(), meta.Limit()) //GetAll recibe el offset (desde q resultado mostrar) y el limit (cuantos desde el offset)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}

		return response.OK("success", allUsers, meta), nil
	}
}

func makeUpdateEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("update status enrollment")

		//Variable con struct de request (datos de atualizacion)
		reqStruct := request.(UpdateRequest)
		//Status debe ir SI O SI y no deb ser vacio
		if reqStruct.Status == nil || *reqStruct.Status == "" {
			return nil, response.BadRequest(ErrStatusRequired.Error())
		}
		if len(*reqStruct.Status) > 2 {
			return nil, response.BadRequest(ErrStatusTooLong.Error())
		}

		id := reqStruct.ID

		err := s.Update(ctx, id, reqStruct.Status)
		if err != nil {
			//Validamos
			if errors.As(err, &ErrEnrollNotFound{}) {
				return nil, response.NotFound(err.Error())
			}
			return nil, response.InternalServerError(err.Error())
		}

		return response.OK("success", reqStruct, nil), nil

	}
}
