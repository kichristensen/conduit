package srv

import (
	"encoding/json"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	"github.com/runconduit/conduit/controller/api/util"
	pb "github.com/runconduit/conduit/controller/gen/public"
	log "github.com/sirupsen/logrus"
)

type (
	jsonError struct {
		Error string `json:"error"`
	}
)

var (
	defaultResourceType = "deployments"
	pbMarshaler         = jsonpb.Marshaler{EmitDefaults: true}
)

func renderJsonError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	log.Error(err.Error())
	rsp, _ := json.Marshal(jsonError{Error: err.Error()})
	w.WriteHeader(status)
	w.Write(rsp)
}

func renderJson(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		renderJsonError(w, err, http.StatusInternalServerError)
		return
	}
	w.Write(jsonResp)
}

func renderJsonPb(w http.ResponseWriter, msg proto.Message) {
	w.Header().Set("Content-Type", "application/json")
	pbMarshaler.Marshal(w, msg)
}

func (h *handler) handleApiVersion(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	version, err := h.apiClient.Version(req.Context(), &pb.Empty{})

	if err != nil {
		renderJsonError(w, err, http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{
		"version": version,
	}
	renderJson(w, resp)
}

func (h *handler) handleApiPods(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	pods, err := h.apiClient.ListPods(req.Context(), &pb.Empty{})
	if err != nil {
		renderJsonError(w, err, http.StatusInternalServerError)
		return
	}

	renderJsonPb(w, pods)
}

func (h *handler) handleApiStat(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	requestParams := util.StatSummaryRequestParams{
		TimeWindow:       req.FormValue("window"),
		ResourceName:     req.FormValue("resource_name"),
		ResourceType:     req.FormValue("resource_type"),
		Namespace:        req.FormValue("namespace"),
		OutToName:        req.FormValue("out_to_name"),
		OutToType:        req.FormValue("out_to_type"),
		OutToNamespace:   req.FormValue("out_to_namespace"),
		OutFromName:      req.FormValue("out_from_name"),
		OutFromType:      req.FormValue("out_from_type"),
		OutFromNamespace: req.FormValue("out_from_namespace"),
	}

	// default to returning deployment stats
	if requestParams.ResourceType == "" {
		requestParams.ResourceType = defaultResourceType
	}

	statRequest, err := util.BuildStatSummaryRequest(requestParams)
	if err != nil {
		renderJsonError(w, err, http.StatusInternalServerError)
		return
	}

	result, err := h.apiClient.StatSummary(req.Context(), statRequest)
	if err != nil {
		renderJsonError(w, err, http.StatusInternalServerError)
		return
	}
	renderJsonPb(w, result)
}
