package enrollment

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IgnacioBO/go_lib_response/response"
)

type (
	//Controller sera una funcion que reciba REspone y Request
	Controller func(ctx context.Context, request interface{}) (interface{}, error)
	Endpoints  struct {
		Create Controller
	}
	//Definiremos una struct para definir el request del Craete, con los campos que quiero recibir y los tags de json
	CreateRequest struct {
		UserID   string `json:"user_id"`
		CourseID string `json:"course_id"`
	}
	//Struct para guardar la cant page por defecto y otras conf
	Config struct {
		LimitPageDefault string
	}
)

// Funcion que se encargará de hacer los endopints
// Para eso necesitaremos una struct que se llamara endpoints
// Esta funcion va a DEVOLVER una struct de Endpoints, estos endpoints son los que vamos a poder utuaizlar en unestro dominio (course)
func MakeEndpoints(s Service, c Config) Endpoints {
	return Endpoints{
		Create: makeCreateEndpoint(s),
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
			return nil, response.InternalServerError(err.Error())
		}

		return response.Created("success", enrollNuevo, nil), nil
	}
}
