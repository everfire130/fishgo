package middleware

import (
	. "github.com/fishedee/app/log"
	. "github.com/fishedee/app/render"
	. "github.com/fishedee/app/router"
	. "github.com/fishedee/app/session"
	. "github.com/fishedee/app/validator"
	. "github.com/fishedee/assert"
	. "github.com/fishedee/encoding"
	. "github.com/fishedee/language"
	"net/http"
	"testing"
)

func a_json(v Validator, s Session) interface{} {
	return v.MustQuery("key")
}

func b_Json(v Validator, s Session) interface{} {
	Throw(10001, "my god")
	return nil
}

func c(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Fish"))
}

func jsonToArray(data string) interface{} {
	var result interface{}
	err := DecodeJson([]byte(data), &result)
	if err != nil {
		panic(err)
	}
	return result
}

func TestEasy(t *testing.T) {
	log, _ := NewLog(LogConfig{Driver: "console"})
	renderFactory, _ := NewRenderFactory(RenderConfig{})
	validatorFactory, _ := NewValidatorFactory(ValidatorConfig{})
	sessionFactory, _ := NewSessionFactory(SessionConfig{Driver: "memory", CookieName: "fishmm"})
	middleware := NewEasyMiddleware(log, validatorFactory, sessionFactory, renderFactory)

	factory := NewRouterFactory()
	factory.Use(NewLogMiddleware(log))
	factory.Use(middleware)
	factory.GET("/a", a_json)
	factory.GET("/b", b_Json)
	factory.GET("/c", c)
	router := factory.Create()

	r, _ := http.NewRequest("GET", "http://example.com/a?key=mmc", nil)
	w := &fakeWriter{}
	router.ServeHttp(w, r)

	AssertEqual(t, jsonToArray(w.Read()), map[string]interface{}{"code": 0.0, "msg": "", "data": "mmc"})

	r2, _ := http.NewRequest("GET", "http://example.com/b?key2=mmc", nil)
	w2 := &fakeWriter{}
	router.ServeHttp(w2, r2)
	AssertEqual(t, jsonToArray(w2.Read()), map[string]interface{}{"code": 10001.0, "msg": "my god"})

	r3, _ := http.NewRequest("GET", "http://example.com/c", nil)
	w3 := &fakeWriter{}
	router.ServeHttp(w3, r3)
	AssertEqual(t, w3.Read(), "Hello Fish")
}