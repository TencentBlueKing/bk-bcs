/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package exec

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/types"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/moby/term"
	"io"
	"runtime"
	"sync"
)

// The default escape key sequence: ctrl-p, ctrl-q
// TODO: This could be moved to `pkg/term`.
var defaultEscapeKeys = []byte{16, 17}

// A hijackedIOStreamer handles copying input to and output from streams to the
// connection.
type hijackedIOStreamer struct {
	streams      Streams
	inputStream  io.ReadCloser
	outputStream io.Writer
	errorStream  io.Writer

	resp types.HijackedResponse

	tty bool
}

// stream handles setting up the IO and then begins streaming stdin/stdout
// to/from the hijacked connection, blocking until it is either done reading
// output, the user inputs the detach key sequence when in TTY mode, or when
// the given context is cancelled.
func (h *hijackedIOStreamer) stream(ctx context.Context) error {
	restoreInput, err := h.setupInput()
	if err != nil {
		return fmt.Errorf("unable to setup input stream: %s", err)
	}

	defer restoreInput()

	outputDone := h.beginOutputStream(restoreInput)
	inputDone, detached := h.beginInputStream(restoreInput)

	select {
	case err := <-outputDone:
		return err
	case <-inputDone:
		// Input stream has closed.
		if h.outputStream != nil || h.errorStream != nil {
			// Wait for output to complete streaming.
			select {
			case err := <-outputDone:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	case err := <-detached:
		// Got a detach key sequence.
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *hijackedIOStreamer) setupInput() (restore func(), err error) {
	if h.inputStream == nil || !h.tty {
		// No need to setup input TTY.
		// The restore func is a nop.
		return func() {}, nil
	}

	if err := setRawTerminal(h.streams); err != nil {
		return nil, fmt.Errorf("unable to set IO streams as raw terminal: %s", err)
	}

	// Use sync.Once so we may call restore multiple times but ensure we
	// only restore the terminal once.
	var restoreOnce sync.Once
	restore = func() {
		restoreOnce.Do(func() {
			restoreTerminal(h.streams, h.inputStream)
		})
	}

	// Wrap the input to detect detach escape sequence.
	// Use default escape keys if an invalid sequence is given.
	escapeKeys := defaultEscapeKeys

	h.inputStream = ioutils.NewReadCloserWrapper(term.NewEscapeProxy(h.inputStream, escapeKeys), h.inputStream.Close)

	return restore, nil
}

func (h *hijackedIOStreamer) beginOutputStream(restoreInput func()) <-chan error {
	if h.outputStream == nil && h.errorStream == nil {
		// There is no need to copy output.
		return nil
	}

	outputDone := make(chan error)
	go func() {
		var err error

		for {
			_, buf, err1 := h.resp.Ws.Conn.ReadMessage()
			if err1 != nil {
				// set err to nil, avoid to always output error when user exit the  exec
				err = nil
				break
			}
			h.outputStream.Write(buf)
		}

		outputDone <- err
	}()

	return outputDone
}

func (h *hijackedIOStreamer) beginInputStream(restoreInput func()) (doneC <-chan struct{}, detachedC <-chan error) {
	inputDone := make(chan struct{})
	detached := make(chan error)

	go func() {
		if h.inputStream != nil {
			//_, err := io.Copy(bufio.NewWriter(h.resp.Conn), h.inputStream)
			_, err := io.Copy(h.resp.Ws, h.inputStream)
			// We should restore the terminal as soon as possible
			// once the connection ends so any following print
			// messages will be in normal type.
			restoreInput()

			blog.Debug("[hijack] End of stdin")

			if _, ok := err.(term.EscapeError); ok {
				detached <- err
				return
			}

			if err != nil {
				// This error will also occur on the receive
				// side (from stdout) where it will be
				// propagated back to the caller.
				blog.Errorf("Error sendStdin: %s", err)
			}
		}

		if err := h.resp.CloseWrite(); err != nil {
			blog.Errorf("Couldn't send EOF: %s", err)
		}

		close(inputDone)
	}()

	return inputDone, detached
}

func setRawTerminal(streams Streams) error {
	if err := streams.In().SetRawTerminal(); err != nil {
		return err
	}
	return streams.Out().SetRawTerminal()
}

// nolint: unparam
func restoreTerminal(streams Streams, in io.Closer) error {
	streams.In().RestoreTerminal()
	streams.Out().RestoreTerminal()
	// WARNING: DO NOT REMOVE THE OS CHECKS !!!
	// For some reason this Close call blocks on darwin..
	// As the client exits right after, simply discard the close
	// until we find a better solution.
	//
	// This can also cause the client on Windows to get stuck in Win32 CloseHandle()
	// in some cases. See https://github.com/docker/docker/issues/28267#issuecomment-288237442
	// Tracked internally at Microsoft by VSO #11352156. In the
	// Windows case, you hit this if you are using the native/v2 console,
	// not the "legacy" console, and you start the client in a new window. eg
	// `start docker run --rm -it microsoft/nanoserver cmd /s /c echo foobar`
	// will hang. Remove start, and it won't repro.
	if in != nil && runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		return in.Close()
	}
	return nil
}
