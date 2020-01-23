// Copyright 2020 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	uriMgmtArtifact   = "/api/management/v1/deployments/artifacts/{id}"
	uriMgmtLink       = uriMgmtArtifact + "/download"
	uriInternalUpload = "/api/management/v1/deployments/tenants/{id}/artifacts"
)

var (
	timeoutSec = 5 * time.Second
)

type Deployments interface {
	GetArtifactLink(ctx context.Context, id, tok string) (string, error)
	DeleteArtifact(ctx context.Context, id, tok string) error
	UploadArtifactInternal(ctx context.Context, path, aid, tid, desc string) error
}

type deployments struct {
	gatewayUrl string
	deplUrl    string
	c          *http.Client
}

func NewDeployments(gatewayUrl, deplUrl string, skipSsl bool) (Deployments, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSsl,
		},
	}

	c := &http.Client{
		Transport: tr,
	}

	return &deployments{
		gatewayUrl: gatewayUrl,
		deplUrl:    deplUrl,
		c:          c,
	}, nil
}

func (d *deployments) GetArtifactLink(ctx context.Context, id, tok string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutSec)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet,
		path.Join(d.gatewayUrl, strings.Replace(uriMgmtLink, "{id}", id, 1)),
		nil)
	if err != nil {
		return "", err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Authorization", "Bearer "+tok)

	res, err := d.c.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.Wrapf(apiErr(res), "failed to get link for artifact %s", id)
	}

	s := struct {
		uri    string
		expire string
	}{}

	err = json.NewDecoder(req.Body).Decode(&s)
	if err != nil {
		return "", errors.Wrapf(err, "failed to decode artifact link for artifact %s", id)
	}

	return s.uri, nil

}

func (d *deployments) DeleteArtifact(ctx context.Context, id, tok string) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutSec)
	defer cancel()

	req, err := http.NewRequest(http.MethodDelete,
		path.Join(d.gatewayUrl, strings.Replace(uriMgmtArtifact, "{id}", id, 1)),
		nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Authorization", "Bearer "+tok)

	res, err := d.c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Wrapf(apiErr(res), "failed to delete artifact %s", id)
	}

	return nil
}

func (d *deployments) UploadArtifactInternal(ctx context.Context, fpath, aid, tid, desc string) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutSec)
	defer cancel()

	artifact, err := os.Open(fpath)
	if err != nil {
		return errors.Wrapf(err, "cannot read artifact file %s", fpath)
	}
	defer artifact.Close()

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	writer.WriteField("id", tid)
	writer.WriteField("artifact_id", aid)
	writer.WriteField("description", desc)

	part, err := writer.CreateFormFile("artifact", filepath.Base(fpath))

	_, err = io.Copy(part, artifact)

	req, err := http.NewRequest(http.MethodPost,
		path.Join(d.deplUrl, strings.Replace(uriInternalUpload, "{id}", tid, 1)),
		body)
	if err != nil {
		return errors.Wrap(err, "cannot create artifact upload request")
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := d.c.Do(req)

	if res.StatusCode != http.StatusCreated {
		return errors.Wrapf(apiErr(res), "failed to upload artifact %s", aid)
	}

	return nil

}

func apiErr(r *http.Response) error {
	e := struct {
		Reqid string `json:"request_id"`
		Msg   string `json:"error"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&e)
	if err != nil {
		return errors.New(fmt.Sprintf("got error code %d from api but failed to decode response", r.StatusCode))
	}

	return errors.New(fmt.Sprintf("http %s reqid %d msg %s ", e.Reqid, r.StatusCode, e.Msg))
}
