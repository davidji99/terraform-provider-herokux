package platform

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/go-resty/resty/v2"
)

type PipelineWebhook struct {
	AppWebhook
	Pipeline heroku.Pipeline `json:"pipeline"`
}

func (p *Platform) CreatePipelineWebhook(ctx context.Context, pipeline string, opts *AppWebhookRequest) (*PipelineWebhook, *resty.Response, error) {
	return p.pipelineWebhook(ctx, http.MethodPost, pipeline, "", opts)
}

func (p *Platform) GetPipelineWebhook(ctx context.Context, pipeline, webhook string) (*PipelineWebhook, *resty.Response, error) {
	return p.pipelineWebhook(ctx, http.MethodGet, pipeline, webhook, nil)
}

func (p *Platform) UpdatePipelineWebhook(ctx context.Context, pipeline, webhook string, changes map[string]interface{}) (*PipelineWebhook, *resty.Response, error) {
	return p.pipelineWebhook(ctx, http.MethodPatch, pipeline, webhook, changes)
}

func (p *Platform) DeletePipelineWebhook(ctx context.Context, pipeline, webhook string) (*resty.Response, error) {
	_, resp, err := p.pipelineWebhook(ctx, http.MethodDelete, pipeline, webhook, nil)
	return resp, err
}

func (p *Platform) pipelineWebhook(ctx context.Context, method, pipeline, webhook string, body interface{}) (*PipelineWebhook, *resty.Response, error) {
	path := p.webhooks.RequestURL("/pipelines/%s/webhooks", url.PathEscape(pipeline))
	if webhook != "" {
		path += "/" + url.PathEscape(webhook)
	}
	var result PipelineWebhook
	resp, err := p.webhooks.R().SetContext(ctx).SetBody(body).SetResult(&result).Execute(method, path)
	if err != nil {
		return nil, resp, err
	}
	if !resp.IsSuccess() {
		// Error bodies can echo signing secrets or receiver authorization headers.
		return nil, resp, fmt.Errorf("%s pipeline webhook: %s", method, resp.Status())
	}
	return &result, resp, nil
}
