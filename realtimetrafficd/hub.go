/*
 * Copyright (C) 2018 Simon Eisenmann
 * Copyright (C) 2014-2017 struktur AG
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"github.com/longsleep/realtimetraffic"
)

type hub struct {
	grabbers    map[string]*realtimetraffic.Grabber
	connections map[*connection]bool
	broadcast   chan *realtimetraffic.Interfacedata
	register    chan *connection
	unregister  chan *connection
}

var h = hub{
	broadcast:   make(chan *realtimetraffic.Interfacedata),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
	grabbers:    make(map[string]*realtimetraffic.Grabber),
}

func (h *hub) run() {
	var eg *realtimetraffic.Grabber
	var ok bool

	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
			if eg, ok = h.grabbers[c.iface]; !ok {
				eg = realtimetraffic.NewGrabber(c.iface)
				h.grabbers[c.iface] = eg
			}
			eg.Start(h.broadcast)
		case c := <-h.unregister:
			delete(h.connections, c)
			close(c.send)
			if eg, ok = h.grabbers[c.iface]; ok {
				eg.Stop()
			}
		case d := <-h.broadcast:
			for c := range h.connections {
				if c.iface == d.Name() {
					if m, err := d.JSON(); err == nil {
						select {
						case c.send <- m:
						default:
							close(c.send)
							delete(h.connections, c)
						}
					}
				}
			}
		}
	}
}
