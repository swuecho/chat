package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/httpx"
	"github.com/swuecho/chat_backend/svc"
	"github.com/swuecho/chat_backend/validation"
)

type ArtifactHandler struct{ service *svc.ArtifactService }

func NewArtifactHandler(service *svc.ArtifactService) *ArtifactHandler {
	return &ArtifactHandler{service: service}
}

type artifactHTTPResponse struct {
	UUID         string    `json:"uuid"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Language     string    `json:"language"`
	MessageUUID  string    `json:"messageUuid"`
	SessionUUID  string    `json:"sessionUuid"`
	SessionTitle string    `json:"sessionTitle"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type artifactPageHTTPResponse struct {
	Items  []artifactHTTPResponse `json:"items"`
	Total  int64                  `json:"total"`
	Limit  int32                  `json:"limit"`
	Offset int32                  `json:"offset"`
}

type updateArtifactRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Language string `json:"language"`
}

func (h *ArtifactHandler) Register(router *mux.Router, registry *apicontract.Registry) {
	security := apicontract.BearerAuth()
	maxPageSize := validation.MaxPageSize
	apicontract.RegisterJSON(router, registry, apicontract.Operation{Method: http.MethodGet, Path: "/artifacts", OperationID: "listArtifacts",
		Summary: "List the current user's artifacts", Tags: []string{"artifacts"}, SuccessStatus: http.StatusOK, Security: security,
		Parameters: []apicontract.Parameter{apicontract.StringQueryParameter("search"), apicontract.StringQueryParameter("type"),
			apicontract.StringQueryParameter("language"), apicontract.StringQueryParameter("sessionUuid"),
			apicontract.IntegerQueryParameter("limit", 1, &maxPageSize), apicontract.IntegerQueryParameter("offset", 0, nil)}}, h.list)
	artifactOperation := func(method, suffix, operationID, summary string) apicontract.Operation {
		return apicontract.Operation{Method: method, Path: "/artifacts/{uuid}" + suffix, OperationID: operationID, Summary: summary,
			Tags: []string{"artifacts"}, SuccessStatus: http.StatusOK, Security: security, Parameters: []apicontract.Parameter{apicontract.UUIDPathParameter("uuid")}}
	}
	apicontract.RegisterJSON(router, registry, artifactOperation(http.MethodPut, "", "updateArtifact", "Update an artifact"), h.update)
	apicontract.RegisterJSON(router, registry, artifactOperation(http.MethodDelete, "", "deleteArtifact", "Delete an artifact"), h.delete)
	apicontract.RegisterJSON(router, registry, artifactOperation(http.MethodPost, "/duplicate", "duplicateArtifact", "Duplicate an artifact"), h.duplicate)
}

func (h *ArtifactHandler) list(r *http.Request, _ apicontract.NoBody) (artifactPageHTTPResponse, error) {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return artifactPageHTTPResponse{}, err
	}
	page, err := httpx.ParsePage(r)
	if err != nil {
		return artifactPageHTTPResponse{}, err
	}
	result, err := h.service.List(r.Context(), svc.ArtifactPageQuery{UserID: userID, Limit: page.Limit, Offset: page.Offset,
		Search: strings.TrimSpace(r.URL.Query().Get("search")), Type: r.URL.Query().Get("type"), Language: r.URL.Query().Get("language"), SessionUUID: r.URL.Query().Get("sessionUuid")})
	if err != nil {
		return artifactPageHTTPResponse{}, dto.WrapError(dto.MapDatabaseError(err), "Failed to list artifacts")
	}
	items := make([]artifactHTTPResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, artifactHTTPResponse{UUID: item.UUID, Type: item.Type, Title: item.Title,
			Content: item.Content, Language: item.Language, MessageUUID: item.MessageUUID, SessionUUID: item.SessionUUID,
			SessionTitle: item.SessionTitle, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt})
	}
	return artifactPageHTTPResponse{Items: items, Total: result.Total, Limit: result.Limit, Offset: result.Offset}, nil
}

func (h *ArtifactHandler) update(r *http.Request, request updateArtifactRequest) (apicontract.NoBody, error) {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return apicontract.NoBody{}, err
	}
	if strings.TrimSpace(request.Title) == "" {
		return apicontract.NoBody{}, domain.Invalid("artifact title is required")
	}
	err = h.service.Update(r.Context(), svc.UpdateArtifactCommand{UUID: mux.Vars(r)["uuid"], Title: request.Title, Content: request.Content, Language: request.Language, UserID: userID})
	if err != nil {
		return apicontract.NoBody{}, dto.WrapError(dto.MapDatabaseError(err), "Failed to update artifact")
	}
	return apicontract.NoBody{}, nil
}

func (h *ArtifactHandler) delete(r *http.Request, _ apicontract.NoBody) (apicontract.NoBody, error) {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return apicontract.NoBody{}, err
	}
	err = h.service.Delete(r.Context(), mux.Vars(r)["uuid"], userID)
	if err != nil {
		return apicontract.NoBody{}, dto.WrapError(dto.MapDatabaseError(err), "Failed to delete artifact")
	}
	return apicontract.NoBody{}, nil
}

func (h *ArtifactHandler) duplicate(r *http.Request, _ apicontract.NoBody) (uuidHTTPResponse, error) {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return uuidHTTPResponse{}, err
	}
	uuid, err := h.service.Duplicate(r.Context(), mux.Vars(r)["uuid"], userID)
	if err != nil {
		return uuidHTTPResponse{}, dto.WrapError(dto.MapDatabaseError(err), "Failed to duplicate artifact")
	}
	return uuidHTTPResponse{UUID: uuid}, nil
}
