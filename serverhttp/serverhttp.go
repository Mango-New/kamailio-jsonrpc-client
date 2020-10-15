package serverhttp

import (
	"encoding/json"
	"net/http"

	"github.com/voipxswitch/kamailio-jsonrpc-client/internal/client"
	"goji.io"
	"goji.io/pat"
)

const (
	requestPath = "/v1/*"
)

type httpHandler struct {
	listenAddr string
	client     client.API
}

// ListenAndServe sets up a new http server
func ListenAndServe(root *goji.Mux, listenAddr string, client client.API) error {
	// setup http mux
	v := goji.SubMux()
	h := httpHandler{
		listenAddr: listenAddr,
		client:     client,
	}
	root.Handle(pat.New(requestPath), v)
	// POST /v1/uac/register returns 200
	v.HandleFunc(pat.Post("/uac/register"), h.register)
	// POST /v1/uac/unregister?domain=test.com&username=1000  returns 200
	v.HandleFunc(pat.Post("/uac/unregister"), h.unregister)
	// POST /v1/uac/list?domain=test.com&username=1000 returns 200
	v.HandleFunc(pat.Post("/uac/list"), h.list)
	return http.ListenAndServe(listenAddr, root)
}

func (h httpHandler) register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	type request struct {
		Username     string `json:"username"`
		Domain       string `json:"domain"`
		AuthUsername string `json:"auth_username"`
		AuthPassword string `json:"auth_password"`
		AuthProxy    string `json:"proxy"`
		RandomDelay  int    `json:"random_delay"`
	}
	z := request{}
	err := json.NewDecoder(r.Body).Decode(&z)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.client.Register(ctx, client.UACAddRequest{
		Username:     z.Username,
		Domain:       z.Domain,
		AuthUsername: z.AuthUsername,
		AuthPassword: z.AuthPassword,
		AuthProxy:    z.AuthProxy,
		RandomDelay:  z.RandomDelay,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func (h httpHandler) unregister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("bad request")
		return
	}
	username := r.Form["username"]
	domain := r.Form["domain"]
	if len(username) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("missing username")
		return
	}
	if len(domain) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("missing domain")
		return
	}
	err = h.client.Unregister(ctx, username[0], domain[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func (h httpHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("bad request")
		return
	}
	domain := r.Form["domain"]
	if len(domain) == 0 {
		x := h.client.ListRegistrations(ctx)
		if len(x) == 0 {
			w.WriteHeader(http.StatusNoContent)
			json.NewEncoder(w).Encode("no registrations")
			return
		}
		json.NewEncoder(w).Encode(x)
		return
	}
	username := r.Form["username"]
	if len(username) == 0 {
		x := h.client.ListRegistrationsByDomain(ctx, domain[0])
		if len(x) == 0 {
			w.WriteHeader(http.StatusNoContent)
			json.NewEncoder(w).Encode("no registrations")
			return
		}
		json.NewEncoder(w).Encode(x)
		return
	}
	x := h.client.ListRegistrationsByUsername(ctx, username[0], domain[0])
	if len(x) == 0 {
		w.WriteHeader(http.StatusNoContent)
		json.NewEncoder(w).Encode("no registrations")
		return
	}
	json.NewEncoder(w).Encode(x)
	return
}
