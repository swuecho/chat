package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/domain"
)

func TestAdaptWritesReturnedError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	Adapt(func(http.ResponseWriter, *http.Request) error {
		return domain.Invalid("bad input")
	})(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"detail":"bad input"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestJSONMarshalsBeforeCommitting(t *testing.T) {
	recorder := httptest.NewRecorder()
	err := JSON(recorder, http.StatusCreated, make(chan int))
	if err == nil {
		t.Fatal("JSON() accepted an unsupported value")
	}
	if recorder.Code != http.StatusOK || recorder.Body.Len() != 0 {
		t.Fatalf("response was committed: status=%d body=%q", recorder.Code, recorder.Body.String())
	}
}

func TestRouteAndPageParsing(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/?limit=25&offset=5", nil)
	request = mux.SetURLVars(request, map[string]string{"id": "7", "uuid": "01990a45-8a36-7e51-bf7c-a8df8d6b8e91"})
	if id, err := Int32Param(request, "id"); err != nil || id != 7 {
		t.Fatalf("Int32Param() = %d, %v", id, err)
	}
	if _, err := UUIDParam(request, "uuid"); err != nil {
		t.Fatal(err)
	}
	page, err := ParsePage(request)
	if err != nil || page.Limit != 25 || page.Offset != 5 {
		t.Fatalf("ParsePage() = %#v, %v", page, err)
	}
}

func TestInvalidHelper(t *testing.T) {
	if !domain.IsKind(Invalid("bad"), domain.KindInvalid) {
		t.Fatal(errors.New("Invalid() did not create a domain validation error"))
	}
}
