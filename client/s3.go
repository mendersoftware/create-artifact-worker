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
	"context"
	"io"
	"net/http"
	"os"
)

type Storage interface {
	Download(ctx context.Context, url, path string) error
}

type storage struct {
	c *http.Client
}

func NewStorage() Storage {
	c := &http.Client{}
	return &storage{
		c: c,
	}
}

func (s *storage) Download(ctx context.Context, url, path string) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutSec)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet,
		url,
		nil)

	req = req.WithContext(ctx)

	res, err := s.c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return err
}
