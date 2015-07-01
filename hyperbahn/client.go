package hyperbahn

// Copyright (c) 2015 Uber Technologies, Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"time"

	"github.com/uber/tchannel/golang"
)

// Client manages Hyperbahn connections and registrations.
type Client struct {
	tchan *tchannel.Channel
	opts  ClientOptions
}

// FailStrategy is the strategy to use when registration fails maxRegistrationFailures
// times consecutively in the background. This is not used if the initial registration fails.
type FailStrategy int

const (
	// FailStrategyFatal will call Fatalf on the channel's logger after triggering handler.OnError.
	// This is the default strategy.
	FailStrategyFatal FailStrategy = iota
	// FailStrategyIgnore will only call handler.OnError, even on fatal errors.
	FailStrategyIgnore
)

const hyperbahnServiceName = "hyperbahn"

// ClientOptions are used to configure this Hyperbahn client.
type ClientOptions struct {
	// Timeout defaults to 1 second if it is not set.
	Timeout      time.Duration
	Handler      Handler
	FailStrategy FailStrategy
}

// NewClient creates a new Hyperbahn client using the given channel.
// config is the environment-specific configuration for Hyperbahn such as the list of initial nodes.
// opts are optional, and are used to customize the client.
func NewClient(ch *tchannel.Channel, config Configuration, opts *ClientOptions) *Client {
	client := &Client{tchan: ch}
	if opts != nil {
		client.opts = *opts
	}
	if client.opts.Timeout == 0 {
		client.opts.Timeout = time.Second
	}
	if client.opts.Handler == nil {
		client.opts.Handler = nullHandler{}
	}

	// Add the given initial nodes as peers.
	for _, node := range config.InitialNodes {
		addPeer(ch, node)
	}

	return client
}

// addPeer adds a peer to the Hyperbahn subchannel.
// TODO(prashant): Start connections to the peers in the background.
func addPeer(ch *tchannel.Channel, hostPort string) {
	peers := ch.GetSubChannel(hyperbahnServiceName).Peers()
	peers.Add(hostPort)
}

// Advertise advertises the service with Hyperbahn, and returns any errors on initial advertisement.
// If the advertisement succeeds, a goroutine is started to re-advertise periodically.
func (c *Client) Advertise() error {
	if err := c.sendAdvertise(); err != nil {
		return err
	}
	c.opts.Handler.On(Advertised)
	go c.advertiseLoop()
	return nil
}