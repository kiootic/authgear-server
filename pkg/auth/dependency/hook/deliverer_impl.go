package hook

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	gohttp "net/http"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/time"

	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/http"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type delivererImpl struct {
	Hooks         *[]config.Hook
	UserConfig    *config.HookUserConfiguration
	AppConfig     *config.HookAppConfiguration
	TimeProvider  time.Provider
	Mutator       Mutator
	NewHTTPClient func() gohttp.Client
}

func NewDeliverer(config *config.TenantConfiguration, timeProvider time.Provider, mutator Mutator) Deliverer {
	return &delivererImpl{
		Hooks:         &config.Hooks,
		UserConfig:    &config.UserConfig.Hook,
		AppConfig:     &config.AppConfig.Hook,
		TimeProvider:  timeProvider,
		Mutator:       mutator,
		NewHTTPClient: func() gohttp.Client { return gohttp.Client{} },
	}
}

func (deliverer *delivererImpl) WillDeliver(eventType event.Type) bool {
	for _, hook := range *deliverer.Hooks {
		if hook.Event == string(eventType) {
			return true
		}
	}
	return false
}

func (deliverer *delivererImpl) DeliverBeforeEvent(e *event.Event, user *model.User) error {
	startTime := deliverer.TimeProvider.Now()
	requestTimeout := gotime.Duration(deliverer.AppConfig.SyncHookTimeout) * gotime.Second
	totalTimeout := gotime.Duration(deliverer.AppConfig.SyncHookTotalTimeout) * gotime.Second

	mutator := deliverer.Mutator.New(e, user)
	client := deliverer.NewHTTPClient()
	client.CheckRedirect = noFollowRedirectPolicy
	client.Timeout = requestTimeout

	for _, hook := range *deliverer.Hooks {
		if hook.Event != string(e.Type) {
			continue
		}

		if deliverer.TimeProvider.Now().Sub(startTime) > totalTimeout {
			return newErrorDeliveryTimeout()
		}

		request, err := deliverer.prepareRequest(hook.URL, e)
		if err != nil {
			return err
		}

		resp, err := performRequest(client, request, true)
		if err != nil {
			return err
		}

		if !resp.IsAllowed {
			return newErrorOperationDisallowed(
				[]OperationDisallowedItem{
					OperationDisallowedItem{
						Reason: resp.Reason,
						Data:   resp.Data,
					},
				},
			)
		}

		if resp.Mutations != nil {
			err = mutator.Add(*resp.Mutations)
			if err != nil {
				return newErrorMutationFailed(err)
			}
		}
	}

	err := mutator.Apply()
	if err != nil {
		return newErrorMutationFailed(err)
	}

	return nil
}

func (deliverer *delivererImpl) DeliverNonBeforeEvent(e *event.Event, timeout gotime.Duration) error {
	client := deliverer.NewHTTPClient()
	client.CheckRedirect = noFollowRedirectPolicy
	client.Timeout = timeout

	for _, hook := range *deliverer.Hooks {
		if hook.Event != string(e.Type) {
			continue
		}

		request, err := deliverer.prepareRequest(hook.URL, e)
		if err != nil {
			return err
		}

		_, err = performRequest(client, request, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (deliverer *delivererImpl) prepareRequest(url string, event *event.Event) (*gohttp.Request, error) {
	body, err := json.Marshal(event)
	if err != nil {
		return nil, newErrorDeliveryFailed(err)
	}

	signature := crypto.HMACSHA256String([]byte(deliverer.UserConfig.Secret), body)

	request, err := gohttp.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, newErrorDeliveryFailed(err)
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add(http.HeaderRequestBodySignature, signature)

	return request, nil
}

func noFollowRedirectPolicy(*gohttp.Request, []*gohttp.Request) error {
	return gohttp.ErrUseLastResponse
}

func performRequest(client gohttp.Client, request *gohttp.Request, withResponse bool) (hookResp *event.HookResponse, err error) {
	var resp *gohttp.Response
	resp, err = client.Do(request)
	if reqError, ok := err.(net.Error); ok && reqError.Timeout() {
		err = newErrorDeliveryTimeout()
		return
	} else if err != nil {
		err = newErrorDeliveryFailed(err)
		return
	}

	defer func() {
		closeError := resp.Body.Close()
		if err == nil && closeError != nil {
			err = newErrorDeliveryFailed(closeError)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = newErrorDeliveryInvalidStatusCode()
		return
	}

	if !withResponse {
		return
	}

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		err = newErrorDeliveryFailed(err)
		return
	}

	hookResp = &event.HookResponse{}
	err = json.Unmarshal(body, &hookResp)
	if err != nil {
		err = newErrorDeliveryFailed(err)
		return
	}

	err = hookResp.Validate()
	if err != nil {
		err = newErrorDeliveryFailed(err)
		return
	}

	return
}
