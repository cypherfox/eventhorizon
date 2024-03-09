// Copyright (c) 2017 - The Event Horizon authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httputils

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	eh "github.com/looplab/eventhorizon"
)

// CommandHandler is a HTTP handler for eventhorizon.Commands. Commands must be
// registered with eventhorizon.RegisterCommand(). It expects a POST with a JSON
// body that will be unmarshaled into the command.
func CommandHandler(commandHandler eh.CommandHandler, commandType eh.CommandType) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "unsupported method: "+r.Method, http.StatusMethodNotAllowed)

			return
		}

		cmd, err := eh.CreateCommand(commandType)
		if err != nil {
			http.Error(w, "could not create command: "+err.Error(), http.StatusBadRequest)

			return
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "could not read command: "+err.Error(), http.StatusBadRequest)

			return
		}

		if err := json.Unmarshal(b, &cmd); err != nil {
			http.Error(w, "could not decode command: "+err.Error(), http.StatusBadRequest)

			return
		}

		// NOTE: Use a new context when handling, else it will be cancelled with
		// the HTTP request which will cause projectors etc to fail if they run
		// async in goroutines past the request.
		ctx := context.Background()
		if err := commandHandler.HandleCommand(ctx, cmd); err != nil {
			http.Error(w, "could not handle command: "+err.Error(), http.StatusBadRequest)

			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

// CommandHandlerWithReply is a HTTP handler for eventhorizon.Commands. Commands must be
// registered with eventhorizon.RegisterCommand(). It expects a POST with a JSON
// body that will be unmarshaled into the command. It differs from CommandHandler by allowing an arbitrary JSON
// document to be returned as the HTTP reply.
func CommandHandlerWithReply(commandHandler eh.CommandHandler, commandType eh.CommandType) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "unsupported method: "+r.Method, http.StatusMethodNotAllowed)

			return
		}

		cmd, err := eh.CreateCommand(commandType)
		if err != nil {
			http.Error(w, "could not create command: "+err.Error(), http.StatusBadRequest)

			return
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "could not read command: "+err.Error(), http.StatusBadRequest)

			return
		}

		if err := json.Unmarshal(b, &cmd); err != nil {
			http.Error(w, "could not decode command: "+err.Error(), http.StatusBadRequest)

			return
		}

		// NOTE: Use a new context when handling, else it will be cancelled with
		// the HTTP request which will cause projectors etc to fail if they run
		// async in goroutines past the request.
		var reply interface{}

		ctx := context.Background()
		if reply, err = commandHandler.HandleCommandWithReply(ctx, cmd); err != nil {
			http.Error(w, "could not handle command: "+err.Error(), http.StatusBadRequest)

			return
		}

		var reply_buf []byte = nil

		if reply != nil {
			reply_buf, err = json.Marshal(reply)
			if err != nil {
				http.Error(w, "could not encode reply: "+err.Error(), http.StatusInternalServerError)

				return
			}
		}

		w.WriteHeader(http.StatusOK)
		if reply_buf != nil {
			w.Write(reply_buf)
		}

	})
}
