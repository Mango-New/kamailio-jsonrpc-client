package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/google/uuid"
	"github.com/romana/rlog"
)

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func handleError(e jsonRPCError) error {
	if e.Code == 0 {
		return nil
	}
	return fmt.Errorf("message [%s] code [%d]", e.Message, e.Code)
}

func (p *API) uaclist(ctx context.Context) []User {
	x := []User{}
	rlog.Debugf("running uac.reg_dump")

	type request struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		ID      string `json:"id"`
	}

	r := request{
		JSONRPC: "2.0",
		Method:  "uac.reg_dump",
		ID:      uuid.New().String(),
	}
	b, err := json.Marshal(&r)
	if err != nil {
		rlog.Errorf("could not marshal [%s]", err.Error())
		return x
	}
	res, err := p.httpClient.Post(p.jsonrpcHTTPAddr, "application/json", bytes.NewBuffer(b))
	if err != nil {
		rlog.Errorf("could not http post [%s]", err.Error())
		return x
	}
	c, err := ioutil.ReadAll(res.Body)
	if err != nil {
		rlog.Errorf("could not read result body [%s]", err.Error())
		return x
	}
	defer res.Body.Close()
	type response struct {
		JSONRPC string `json:"jsonrpc"`
		Result  []struct {
			UUID         string `json:"l_uuid"`
			LUsername    string `json:"l_username"`
			LDomain      string `json:"l_domain"`
			RUsername    string `json:"r_username"`
			RDomain      string `json:"r_domain"`
			Realm        string `json:"realm"`
			AuthUsernme  string `json:"auth_username"`
			AuthPassword string `json:"auth_password"`
			AuthHA1      string `json:"auth_ha1"`
			AuthProxy    string `json:"auth_proxy"`
			Expires      int    `json:"expires"`
			Flags        int    `json:"flags"`
			RegDelay     int    `json:"reg_delay"`
			Socket       string `json:"socket"`
		} `json:"result"`
		ID string `json:"id"`
	}
	z := response{}
	if err = json.Unmarshal(c, &z); err != nil {
		rlog.Errorf("could not unmarshal [%s]", err.Error())
		return x
	}
	for _, v := range z.Result {
		j := User{UUID: v.UUID, Username: v.LUsername, Domain: v.LDomain, Expires: v.Expires, RegStatus: "unregistered"}
		if v.Flags == 20 {
			j.RegStatus = "registered"
		} else if v.Flags == 16 {
			j.RegStatus = "trying"
		} else {
			rlog.Infof("unmatched flag [%d] for user [%s@%s]", v.Flags, v.LUsername, v.LDomain)
		}
		x = append(x, j)
	}
	return x
}

func (p *API) uacRemove(ctx context.Context, id string) error {
	rlog.Debugf("removing registation with uuid [%s]", id)
	type params struct {
		UUID string `json:"l_uuid"`
	}

	type request struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  params `json:"params"`
		ID      string `json:"id"`
	}

	r := request{
		JSONRPC: "2.0",
		Method:  "uac.reg_remove",
		ID:      uuid.New().String(),
		Params: params{
			UUID: id,
		},
	}
	b, err := json.Marshal(&r)
	if err != nil {
		return err
	}
	res, err := p.httpClient.Post(p.jsonrpcHTTPAddr, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	x, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	type response struct {
		JSONRPC string       `json:"jsonrpc"`
		Error   jsonRPCError `json:"error,omitempty"`
		ID      string       `json:"id"`
	}
	z := response{}
	if err = json.Unmarshal(x, &z); err != nil {
		return err
	}
	return handleError(z.Error)
}

func (p *API) uacAdd(ctx context.Context, id string, username string, domain string, authUsername string, authPassword string, authProxy string, expires int, regDelay int) error {
	rlog.Debugf("adding registation with uuid [%s]", id)
	type params struct {
		UUID         string `json:"l_uuid"`
		Username     string `json:"l_username"`
		LDomain      string `json:"l_domain"`
		RUsername    string `json:"r_username"`
		RDomain      string `json:"r_domain"`
		Realm        string `json:"realm"`
		AuthUsernme  string `json:"auth_username"`
		AuthPassword string `json:"auth_password"`
		AuthHA1      string `json:"auth_ha1"`
		AuthProxy    string `json:"auth_proxy"`
		Expires      int    `json:"expires"`
		Flags        int    `json:"flags"`
		RegDelay     int    `json:"reg_delay"`
		Socket       string `json:"socket"`
	}

	type request struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  params `json:"params"`
		ID      string `json:"id"`
	}

	r := request{
		JSONRPC: "2.0",
		Method:  "uac.reg_add",
		ID:      uuid.New().String(),
		Params: params{
			UUID:         id,
			Username:     username,
			LDomain:      domain,
			RUsername:    username,
			RDomain:      domain,
			Realm:        domain,
			AuthUsernme:  authUsername,
			AuthPassword: authPassword,
			AuthProxy:    authProxy,
			Expires:      expires,
			RegDelay:     regDelay,
		},
	}
	b, err := json.Marshal(&r)
	if err != nil {
		return err
	}
	res, err := p.httpClient.Post(p.jsonrpcHTTPAddr, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	x, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	type response struct {
		JSONRPC string       `json:"jsonrpc"`
		Error   jsonRPCError `json:"error,omitempty"`
		ID      string       `json:"id"`
	}
	z := response{}
	if err = json.Unmarshal(x, &z); err != nil {
		return err
	}
	return handleError(z.Error)
}
