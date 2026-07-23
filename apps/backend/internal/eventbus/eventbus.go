// Package eventbus re-exports packages/event for backward compatibility.
package eventbus

import "github.com/aistudio/packages/event"

type Topic = event.Topic

type EventHandler = event.EventHandler

type Subscription = event.Subscription

type Event = event.Event

type EventBus = event.EventBus

type Option = event.Option

var New = event.New
var WithHistorySize = event.WithHistorySize
var WithTrace = event.WithTrace